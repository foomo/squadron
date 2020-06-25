package actions

import (
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func ValidatePath(wd string, p *string) error {
	if !filepath.IsAbs(*p) {
		*p = path.Join(wd, *p)
	}
	absPath, err := filepath.Abs(*p)
	if err != nil {
		return err
	}
	_, err = os.Stat(absPath)
	if err != nil {
		return err
	}
	*p = absPath
	return nil
}

func CheckIngressController(l *logrus.Entry, name string) error {
	// pods, err := getPodsByLabels(l, []string{fmt.Sprintf("app.kubernetes.io/name=%v", name)})
	// if err != nil {
	// 	return fmt.Errorf("error while checking ingress controller %q: %s", name, err)
	// }
	// if len(pods) == 0 {
	// 	return fmt.Errorf("ingress controller %q not present on any namespace", name)
	// }
	return nil
	// todo this uses kube from grapple
}
