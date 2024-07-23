package models

type Queue string

const (
	TemplateMessageQueue Queue = "template-message"
)

type Exchange string

const (
	TemplateExchangeName Exchange = "template-exchange"
)
