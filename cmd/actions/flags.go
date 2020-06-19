package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/foomo/configurd"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type HostPort struct {
	Host string
	Port int
}

func newHostPort(host string, port int) *HostPort {
	addr, err := configurd.CheckTCPConnection(host, port)
	if err == nil {
		host = addr.IP.String()
		port = addr.Port
	}
	return &HostPort{host, port}
}

func (lf *HostPort) Set(value string) error {
	pieces := strings.Split(value, ":")
	if pieces[0] != "" {
		lf.Host = pieces[0]
	}
	var err error
	if len(pieces) == 2 && pieces[1] != "" {
		lf.Port, err = strconv.Atoi(pieces[1])
	}
	if err != nil {
		return err
	}
	addr, err := configurd.CheckTCPConnection(lf.Host, lf.Port)
	if err != nil {
		return err
	}
	lf.Host = addr.IP.String()
	lf.Port = addr.Port
	return nil
}

func (lf HostPort) String() string {
	return fmt.Sprintf("%v:%v", lf.Host, lf.Port)
}

func (*HostPort) Type() string {
	return "host:port"
}

type KubeResource struct {
	name string
	l    *logrus.Entry
}

func newKubeResource(name string) *KubeResource {
	return &KubeResource{name, newLogger(false)}
}

func (*KubeResource) Type() string {
	return "string"
}

func (kr KubeResource) String() string {
	return kr.name
}

func (kr *KubeResource) Set(value string) error {
	kr.name = value
	return nil
}

type Namespace struct {
	*KubeResource
}

func newNamespace(defaultValue string) *Namespace {
	return &Namespace{newKubeResource(defaultValue)}
}

func (n *Namespace) Validate() error {
	return configurd.ValidateNamespace(n.l, n.name)
}

type Deployment struct {
	*KubeResource
	ns  *Namespace
	obj *appsv1.Deployment
}

func newDeployment(ns *Namespace) *Deployment {
	return &Deployment{newKubeResource(""), ns, nil}
}

func (d *Deployment) Validate() error {
	var err error
	if err := configurd.ValidateDeployment(d.l, d.ns.name, d.name); err != nil {
		return err
	}
	d.obj, err = configurd.GetDeployment(d.l, d.ns.name, d.name)
	if err != nil {
		return err
	}
	if d.obj == nil {
		return fmt.Errorf("couldnt get deployment resource object")
	}
	return nil
}

func (d *Deployment) Resource() *appsv1.Deployment {
	return d.obj
}

type Pod struct {
	*KubeResource
	d *Deployment
}

func newPod(d *Deployment) *Pod {
	return &Pod{newKubeResource(""), d}
}

func (p *Pod) Validate() error {
	if p.name == "" {
		var err error
		p.name, err = configurd.GetMostRecentPodBySelectors(p.l, p.d.obj.Spec.Selector.MatchLabels, p.d.ns.name)
		if err != nil || p.name == "" {
			return err
		}
		return nil
	}
	if err := configurd.ValidatePod(p.l, p.d.Resource(), p.name); err != nil {
		return err
	}
	return nil
}

type Container struct {
	*KubeResource
	d   *Deployment
	obj *corev1.Container
}

func newContainer(d *Deployment) *Container {
	return &Container{newKubeResource(""), d, nil}
}

func (c *Container) Validate() error {
	if c.name == "" {
		c.name = c.d.name
	}
	if err := configurd.ValidateContainer(c.l, c.d.Resource(), c.name); err != nil {
		return err
	}
	for _, container := range c.d.obj.Spec.Template.Spec.Containers {
		if c.name == container.Name {
			c.obj = &container
		}
	}
	if c.obj == nil {
		return fmt.Errorf("couldnt get container resource object")
	}
	return nil
}

func (c *Container) getImage() string {
	return strings.Split(c.obj.Image, ":")[0]
}

func (c *Container) getTag() string {
	return strings.Split(c.obj.Image, ":")[1]
}

type Path string

func newPath() Path {
	wd, _ := os.Getwd()
	return Path(wd)
}

func (p *Path) Set(value string) error {
	absPath, err := filepath.Abs(value)
	if err != nil {
		return err
	}
	_, err = os.Stat(absPath)
	if err != nil {
		return err
	}
	*p = Path(absPath)
	return nil
}

func (p Path) String() string {
	return string(p)
}

func (*Path) Type() string {
	return "path"
}

type StringList struct {
	separator string
	items     []string
}

func newStringList(separator string) *StringList {
	return &StringList{separator: separator}
}

func (sl *StringList) Set(value string) error {
	if value == "" {
		return nil
	}
	sl.items = strings.Split(value, sl.separator)
	return nil
}

func (sl StringList) String() string {
	return strings.Join(sl.items, sl.separator)
}

func (sl *StringList) Type() string {
	return fmt.Sprintf("string separated: %q", sl.separator)
}
