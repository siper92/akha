package types

type AIBackend[p string, ans string] interface {
	Ask(prompt p) (ans, error)
	Rate(prompt p, answer ans) (int, error)
	RateText(prompt p, answer ans) (string, error)
}

type APIAIBackend[p string, ans string] interface {
	AIBackend[p, ans]
	Init(
		service APIService,
		dbService DbService,
		loggerService LoggerService,
	) error
}

type AIFrontend[p string, ans string] interface {
	AIBackend[p, ans]
}

type APIAIFrontend[p string, ans string] interface {
	AIFrontend[p, ans]
	Init(
		service APIServiceFrontend,
		dbService DbService,
		loggerService LoggerService,
	) error
}

type AIGenerator interface {
	Init(
		be AIBackend[string, string],
		fe AIFrontend[string, string],
	) error
}
