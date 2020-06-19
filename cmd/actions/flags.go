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

func (n *Namespace) Set(value string) error {
	return configurd.ValidateNamespace(n.l, n.name)
}

type Deployment struct {
	*KubeResource
	ns  *Namespace
	res *appsv1.Deployment
}

func newDeployment(ns *Namespace) *Deployment {
	return &Deployment{newKubeResource(""), ns, nil}
}

func (d *Deployment) Set(value string) error {
	return configurd.ValidateDeployment(d.l, d.ns.name, value)
}

func (d *Deployment) Resource() (*appsv1.Deployment, error) {
	if d.res == nil {
		res, err := configurd.GetDeployment(d.l, d.ns.name, d.name)
		if err != nil {
			return nil, err
		}
		d.res = res
	}
	return d.res, nil
}

type Pod struct {
	*KubeResource
	d *Deployment
}

func newPod(d *Deployment) *Pod {
	return &Pod{newKubeResource(""), d}
}

func (p *Pod) Set(value string) error {
	p.name = value
	res, err := p.d.Resource()
	if err != nil {
		return err
	}
	return configurd.ValidatePod(p.l, res, &p.name)
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
	c.name = value
	res, err := c.d.Resource()
	if err != nil {
		return err
	}
	return configurd.ValidateContainer(c.l, res, &c.name)
}

func (c *Container) ValidateImage(image, tag *string) error {
	if *image == "" {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if c.name == container.Name {
				pieces := strings.Split(container.Image, ":")
				if len(pieces) != 2 {
					return fmt.Errorf("deployment image %q has invalid format", container.Image)
				}
				*image = pieces[0]
				*tag = pieces[1]
				return nil
			}
		}
	}
	return fmt.Errorf("couldnt find deployment %q image for container %q", c.d.name, c.name)
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
