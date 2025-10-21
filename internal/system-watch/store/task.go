package store

import (
	"context"
	"errors"
	"gorm.io/gorm"

	"github.com/pachirode/distributed-task-system-demo/internal/system-watch/model"
	"github.com/pachirode/distributed-task-system-demo/pkg/meta"
)

type ITaskStore interface {
	Create(ctx context.Context, task *model.TaskM) error
	Get(ctx context.Context, taskID string) (*model.TaskM, error)
	List(ctx context.Context, opts ...meta.GetOptions) (int64, []*model.TaskM, error)
	Update(ctx context.Context, task *model.TaskM) error
	Delete(ctx context.Context, taskID string) error
}

type taskStore struct {
	ds *datastore
}

func newTaskMStore(ds *datastore) *taskStore {
	return &taskStore{ds}
}

var _ ITaskStore = (*taskStore)(nil)

func (d *taskStore) db(ctx context.Context) *gorm.DB {
	return d.ds.DB(ctx)
}

func (d *taskStore) Create(ctx context.Context, task *model.TaskM) error {
	return d.db(ctx).Create(&task).Error
}

func (d *taskStore) Get(ctx context.Context, taskID string) (*model.TaskM, error) {
	task := &model.TaskM{}
	if err := d.db(ctx).Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

func (d *taskStore) List(ctx context.Context, opts ...meta.GetOptions) (count int64, ret []*model.TaskM, err error) {
	o := meta.NewListOptions(opts...)

	ans := d.db(ctx).
		Where(o.Filters).
		Not(o.Not).
		Offset(o.Offset).
		Limit(defaultLimit(o.Limit)).
		Order(defaultOrder(o.Order)).
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count)

	return count, ret, ans.Error
}

func (d *taskStore) Update(ctx context.Context, task *model.TaskM) error {
	return d.db(ctx).Save(task).Error
}

func (d *taskStore) Delete(ctx context.Context, taskID string) error {
	err := d.db(ctx).Where("id = ?", taskID).Delete(&model.TaskM{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}
