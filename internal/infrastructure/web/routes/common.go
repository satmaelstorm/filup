package routes

const (
	Metrics = "/metrics"

	Upload       = "/upload"
	StartUpload  = Upload + "/start"
	FinishUpload = Upload + "/finish"
	UploadPart   = Upload + "/part"
)
