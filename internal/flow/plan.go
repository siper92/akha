package flow

type ExecutionPlan interface {
	String() string
	Plan() PlanSchema
}

type PlanSchema interface {
}
