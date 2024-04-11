package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// ReadJob reads a Job from Kubernetes.
func (client *Client) ReadJob(name string) (*models.JobPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s job %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when reading k8s job. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.BatchV1().Jobs(client.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateJobPublicFromRead(res), nil
}

// CreateJob creates a Job in Kubernetes.
func (client *Client) CreateJob(public *models.JobPublic) (*models.JobPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s job %s. details: %w", public.Name, err)
	}

	job, err := client.K8sClient.BatchV1().Jobs(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateJobPublicFromRead(job), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateJobManifest(public)
	res, err := client.K8sClient.BatchV1().Jobs(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateJobPublicFromRead(res), nil
}

// DeleteJob deletes a Job in Kubernetes.
func (client *Client) DeleteJob(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s job %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting k8s job. Assuming it was deleted")
		return nil
	}

	err := client.K8sClient.BatchV1().Jobs(client.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &[]metav1.DeletionPropagation{metav1.DeletePropagationBackground}[0],
	})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}

// CreateOneShotJob creates a Job in Kubernetes that runs once and then deletes itself.
//
// This is useful for running tasks, such as creating NFS PVs, that should only be run once.
//
// This function is blocking.
func (client *Client) CreateOneShotJob(public *models.JobPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create one-shot k8s job %s. details: %w", public.Name, err)
	}

	_, err := client.CreateJob(public)
	if err != nil {
		return makeError(err)
	}

	// Wait for the job to complete.
	maxIter := 60
	iter := 0
	for {
		k8sJob, err := client.K8sClient.BatchV1().Jobs(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
		if err != nil {
			if IsNotFoundErr(err) {
				return makeError(fmt.Errorf("job %s was deleted before it could complete", public.Name))
			}

			return makeError(err)
		}

		if k8sJob.Status.Succeeded > 0 {
			break
		}

		time.Sleep(1 * time.Second)

		iter++
		if iter > maxIter {
			return makeError(fmt.Errorf("job %s did not complete in time", public.Name))
		}
	}

	// Delete the job.
	err = client.K8sClient.BatchV1().Jobs(public.Namespace).Delete(context.TODO(), public.Name, metav1.DeleteOptions{
		PropagationPolicy: &[]metav1.DeletionPropagation{metav1.DeletePropagationBackground}[0],
	})

	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}
