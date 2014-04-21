package remote

type Operater interface {
	GetBucketOperater() BucketOperater
	GetObjectOperater() ObjectOperater
}

type BucketOperater interface {
	List() (ListAllMyBucketLister, error)
	Put(name string) error
	Delete(name string) error
}

type ObjectOperater interface {
	List(remoteDir, marker, maxKeys string) (ObjectLister, error)
	ListAll(remoteDir string) (ObjectLister, error)
	Put(remoteDir, localDir, filename string) error
	Get(remoteDir, localDir, objectName string) error
	Delete(remoteDir, objectName string) error
}
// ListAllMyBucketList
type ListAllMyBucketLister struct {
	Owner      Owner
	BucketList []Bucketer
}

type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

type Bucketer struct {
	Location     string
	Name         string
	Acl          string
	CreationDate string
}
// end ListAllMyBucketList

// ListObjectList
type ObjectLister struct {
	Name        string
	Prefix      string
	Marker      string
	MaxKeys     int
	Delimiter   string
	IsTruncated bool
	ObjectList  []Objecter
}

type Objecter struct {
	Key          string
	LastModified string
	ETag         string
	Type         string
	Size         int
	StorageClass string
	Owner        Owner 
}
// end ListObjectList