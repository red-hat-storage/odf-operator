package util

const (
	IndexBucketName              = "index:bucketName"
	StorageClassDriverNamePrefix = "openshift-storage"
	RbdDriverNameSuffix          = ".rbd.csi.ceph.com"
	CephFSDriverNameSuffix       = ".cephfs.csi.ceph.com"
	ObcDriverNameSuffix          = ".ceph.rook.io/bucket"
	RbdDriverName                = StorageClassDriverNamePrefix + RbdDriverNameSuffix
	CephFSDriverName             = StorageClassDriverNamePrefix + CephFSDriverNameSuffix
	ObcDriverName                = StorageClassDriverNamePrefix + ObcDriverNameSuffix
	NoobaaDriverNameSuffix       = ".noobaa.io/obc"
	NoobaaResourceName           = "noobaa"
	PhaseIgnored                 = "Ignored"
)
