package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type (
	Message struct {
		Status string
		Detail string
	}

	Deployment struct {
		Namespace  string
		Deployment string
		Container  string
		Image      string
	}

	KubeBackend struct {
		KubeConfig    *rest.Config
		KubeClientSet *kubernetes.Clientset
	}
)

func getPingPong(c echo.Context) error {
	message := &Message{
		Status: "okay",
		Detail: "server is running",
	}
	return c.JSON(http.StatusOK, message)
}

func (b KubeBackend) getDeployment(c echo.Context) error {
	deploymentsClient := b.KubeClientSet.AppsV1().Deployments(c.Param("namespace"))
	result, err := deploymentsClient.Get(context.TODO(), c.Param("deployment"), metav1.GetOptions{})
	if err != nil {
		fmt.Println("error getting deployment details")
		message := &Message{
			Status: "failed",
			Detail: "could not find deployment",
		}
		return c.JSON(http.StatusNotFound, message)
	}
	image := ""
	for i, ctr := range result.Spec.Template.Spec.Containers {
		fmt.Printf("index = %d, name = %s\n", i, ctr.Name)
		if ctr.Name == c.Param("container") {
			fmt.Printf("found the container\n")
			image = ctr.Image
		}
	}
	deployment := &Deployment{
		Namespace:  c.Param("namespace"),
		Deployment: c.Param("deployment"),
		Container:  c.Param("container"),
		Image:      image,
	}
	return c.JSON(http.StatusOK, deployment)
}

func (b KubeBackend) setDeploymentImage(c echo.Context) error {
	d := &Deployment{}
	if err := c.Bind(d); err != nil {
		return err
	}
	if d.Image == "" {
		message := &Message{
			Status: "failed",
			Detail: "no image specified in request",
		}
		return c.JSON(http.StatusBadRequest, message)
	}
	deploymentsClient := b.KubeClientSet.AppsV1().Deployments(c.Param("namespace"))
	result, err := deploymentsClient.Get(context.TODO(), c.Param("deployment"), metav1.GetOptions{})
	if err != nil {
		fmt.Println("error getting deployment details")
		message := &Message{
			Status: "failed",
			Detail: "could not find deployment",
		}
		return c.JSON(http.StatusNotFound, message)
	}
	// need to find index for container
	ctrIndex := -1
	for i, ctr := range result.Spec.Template.Spec.Containers {
		if ctr.Name == c.Param("container") {
			ctrIndex = i
		}
	}
	if ctrIndex == -1 {
		message := &Message{
			Status: "failed",
			Detail: "could not find container in deployment",
		}
		return c.JSON(http.StatusNotFound, message)
	}
	result.Spec.Template.Spec.Containers[ctrIndex].Image = d.Image
	_, updateErr := deploymentsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		message := &Message{
			Status: "failed",
			Detail: "update of deployment failed",
		}
		return c.JSON(http.StatusInternalServerError, message)
	}
	deployment := &Deployment{
		Namespace:  c.Param("namespace"),
		Deployment: c.Param("deployment"),
		Container:  c.Param("container"),
		Image:      d.Image,
	}
	return c.JSON(http.StatusOK, deployment)
}

func InitKubeBackend() *KubeBackend {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolutepath to kubeconfig file")
	}
	flag.Parse()

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	kubeClientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	backend := &KubeBackend{
		KubeConfig:    kubeConfig,
		KubeClientSet: kubeClientSet,
	}

	return backend
}

func main() {

	b := InitKubeBackend()

	pods, err := b.KubeClientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There %d pods in the cluster\n", len(pods.Items))

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", getPingPong)
	e.GET("/deployment/:namespace/:deployment/:container", b.getDeployment)
	e.POST("/deployment/:namespace/:deployment/:container", b.setDeploymentImage)

	e.Logger.Fatal(e.Start(":8080"))
}
