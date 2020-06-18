package actions

import (
	"flag"
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
	return &HostPort{host, port}
}

func (lf *HostPort) Set(value string) error {
	pieces := strings.Split(value, ":")
	if len(pieces) != 2 || pieces[1] == "" {
		return fmt.Errorf("unable to parse %q", value)
	}
	if pieces[0] != "" {
		lf.Host = pieces[0]
	}
	var err error
	lf.Port, err = strconv.Atoi(pieces[1])
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

func getValue(flag flag.Value) string {
	if flag.String() == "" {
		flag.Set("")
	}
	return flag.String()
}

type Namespace struct {
	*KubeResource
}

func newNamespace(defaultValue string) *Namespace {
	return &Namespace{newKubeResource(defaultValue)}
}

func (n *Namespace) Set(value string) error {
	if value == "" {
		return nil
	}
	err := configurd.ValidateNamespace(n.l, value)
	if err != nil {
		return err
	}
	n.name = value
	return nil
}

func (ns *Namespace) Value() string {
	return getValue(ns)
}

type Deployment struct {
	*KubeResource
	ns  *Namespace
	obj *appsv1.Deployment
}

func newDeployment(ns *Namespace) *Deployment {
	return &Deployment{newKubeResource(""), ns, nil}
}

func (d *Deployment) Set(value string) error {
	err := configurd.ValidateDeployment(d.l, d.ns.name, value)
	if err != nil {
		return err
	}
	d.name = value
	d.obj, err = configurd.GetDeployment(d.l, d.ns.name, d.name)
	if err != nil {
		return err
	}
	return nil
}

func (d *Deployment) Resource() *appsv1.Deployment {
	return d.obj
}

func (d *Deployment) Value() string {
	return getValue(d)
}

type Pod struct {
	*KubeResource
	d *Deployment
}

func newPod(d *Deployment) *Pod {
	return &Pod{newKubeResource(""), d}
}

func (p *Pod) Set(value string) error {
	if value == "" {
		pod, err := configurd.GetMostRecentPodBySelectors(p.l, p.d.obj.Spec.Selector.MatchLabels, p.d.ns.name)
		if err != nil || pod == "" {
			return err
		}
		p.name = pod

		return nil
	}
	err := configurd.ValidatePod(p.l, p.d.Resource(), value)
	if err != nil {
		return err
	}
	p.name = value
	return nil
}

func (p *Pod) Value() string {
	return getValue(p)
}

type Container struct {
	*KubeResource
	d   *Deployment
	obj *corev1.Container
}

func newContainer(d *Deployment) *Container {
	return &Container{newKubeResource(""), d, nil}
}

func (c *Container) Set(value string) error {
	if value == "" {
		value = c.d.name
	}
	err := configurd.ValidateContainer(c.l, c.d.Resource(), value)
	if err != nil {
		return err
	}
	c.name = value
	for _, container := range c.d.obj.Spec.Template.Spec.Containers {
		if c.name == container.Name {
			c.obj = &container
		}
	}
	return nil
}

func (c *Container) Value() string {
	return getValue(c)
}

func (c *Container) getImage() string {
	return strings.Split(c.obj.Image, ":")[0]
}

func (c *Container) getTag() string {
	return strings.Split(c.obj.Image, ":")[1]
}

type Path string

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
