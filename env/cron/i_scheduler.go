package cron

import (
	"github.com/gorhill/cronexpr"
	"github.com/ottemo/foundation/env"
	"time"
)

// ListTasks returns a list of task names currently available
func (it *DefaultCronScheduler) ListTasks() []string {
	var result []string
	for taskName := range it.tasks {
		result = append(result, taskName)
	}
	return result
}

// RegisterTask registers a new task routine by a given task name
//   - returns error no non unique name
func (it *DefaultCronScheduler) RegisterTask(name string, task env.FuncCronTask) error {
	if _, present := it.tasks[name]; present {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "231fa82d-c357-498d-b0b3-f4daee7e25c5", "task already exists")
	}

	it.tasks[name] = task

	return nil
}

// ScheduleOnce schedules task execution with a given params
func (it *DefaultCronScheduler) ScheduleAtTime(scheduleTime time.Time, taskName string, params map[string]interface{}) (env.InterfaceSchedule, error) {

	task, present := it.tasks[taskName]
	if !present {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ee521c4f-b84c-4238-bdac-ce61a37267a3", "unexistent task")
	}

	schedule := &DefaultCronSchedule{
		CronExpr:  "",
		TaskName:  taskName,
		Params:    params,
		Repeat:    false,
		Time:      scheduleTime,
		task:      task,
		expr:      nil,
		scheduler: it}

	it.schedules = append(it.schedules, schedule)

	go schedule.Execute()

	return schedule, nil
}

// ScheduleOnce schedules task execution with a given params
func (it *DefaultCronScheduler) ScheduleOnce(cronExpr string, taskName string, params map[string]interface{}) (env.InterfaceSchedule, error) {
	expr, err := cronexpr.Parse(cronExpr)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	task, present := it.tasks[taskName]
	if !present {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ee521c4f-b84c-4238-bdac-ce61a37267a3", "unexistent task")
	}

	schedule := &DefaultCronSchedule{
		CronExpr:  cronExpr,
		TaskName:  taskName,
		Params:    params,
		Repeat:    false,
		task:      task,
		expr:      expr,
		scheduler: it}

	it.schedules = append(it.schedules, schedule)

	go schedule.Execute()

	return schedule, nil
}

// ScheduleRepeat schedules task execution with a given params
func (it *DefaultCronScheduler) ScheduleRepeat(cronExpr string, taskName string, params map[string]interface{}) (env.InterfaceSchedule, error) {

	expr, err := cronexpr.Parse(cronExpr)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	task, present := it.tasks[taskName]
	if !present {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ee521c4f-b84c-4238-bdac-ce61a37267a3", "unexistent task")
	}

	schedule := &DefaultCronSchedule{
		CronExpr:  cronExpr,
		TaskName:  taskName,
		Params:    params,
		Repeat:    true,
		task:      task,
		expr:      expr,
		scheduler: it}

	it.schedules = append(it.schedules, schedule)

	go schedule.Execute()

	return schedule, nil
}

// ListSchedules returns list of currently registered schedules
func (it *DefaultCronScheduler) ListSchedules() []env.InterfaceSchedule {
	var result []env.InterfaceSchedule
	for _, item := range it.schedules {
		result = append(result, item)
	}
	return result
}
