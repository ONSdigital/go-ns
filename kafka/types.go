package kafka

/* PublishMessage is not used externally yet - message is within publish-scheduler
type PublishMessage struct {
	ScheduleId     int64
	CollectionId   string
	CollectionPath string
	EncryptionKey  string
	ScheduleTime   int64
	Files          []FileResource
	FilesToDelete  []FileResource
}
*/

type FileResource struct {
	Id       int64  // in DB
	Location string // e.g. "s3://..."
	Uri      string // on website
}

type ScheduleMessage struct {
	Action         string
	CollectionId   string
	CollectionPath string
	ScheduleTime   string
	Files          []FileResource
	UrisToDelete   []string
}

type PublishFileMessage struct {
	ScheduleId     int64
	FileId         int64
	CollectionId   string
	CollectionPath string
	EncryptionKey  string
	FileLocation   string
	Uri            string
}

type PublishDeleteMessage struct {
	ScheduleId   int64
	DeleteId     int64
	CollectionId string
	Uri          string
}

// S3Location and FileContent are mutually exclusive
type FileCompleteMessage struct {
	ScheduleId   int64
	FileId       int64
	CollectionId string
	Uri          string
	S3Location   string
	FileContent  string
}

// (FileId) and (DeleteId) are mutually exclusive
type FileCompleteFlagMessage struct {
	ScheduleId   int64
	CollectionId string
	FileId       int64
	Uri          string
	DeleteId     int64
}

type CollectionCompleteMessage struct {
	ScheduleId   int64
	CollectionId string
}
