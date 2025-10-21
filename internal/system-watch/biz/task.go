package biz

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachirode/distributed-task-system-demo/internal/system-watch/model"
)

func toJob(task *model.TaskM) *batchv1.Job {
	backoffLimit := int32(1)
	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: task.Name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "task-job",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "task",
							Image:           task.Info.Image,
							Command:         task.Info.Command,
							Args:            task.Info.Args,
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backoffLimit,
		},
	}
	return jobSpec
}

func toTaskStatus(job *batchv1.Job) string {
	conditions := job.Status.Conditions
	switch {
	case isJobSuspended(job):
		return model.TaskStatusPending
	case isJobActive(job):
		return model.TaskStatusRunning
	case isConditionTrue(conditions, batchv1.JobComplete):
		return model.TaskStatusSucceeded
	case isConditionTrue(conditions, batchv1.JobFailed):
		return model.TaskStatusFailed
	case isConditionTrue(conditions, batchv1.JobFailureTarget):
		return model.TaskStatusFailed
	case isConditionTrue(conditions, batchv1.JobSuccessCriteriaMet):
		return model.TaskStatusSucceeded
	default:
		return model.TaskStatusUnknown
	}
}
func isJobSuspended(job *batchv1.Job) bool {
	return job.Spec.Suspend != nil && *job.Spec.Suspend
}

func isJobActive(job *batchv1.Job) bool {
	return job.Status.Active > 0
}

func isConditionTrue(conditions []batchv1.JobCondition, conditionType batchv1.JobConditionType) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
