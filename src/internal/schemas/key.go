package schemas

import "time"

// Permission define o nível de acesso da chave: "read" ou "readwrite"
type KeyPermission string

const (
	KeyPermissionRead      KeyPermission = "read"
	KeyPermissionReadWrite KeyPermission = "readwrite"
)

type Key struct {
	ID         uint          `gorm:"primaryKey"`
	UserID     string        `gorm:"index;not null"`
	Key        string        `gorm:"uniqueIndex;not null"`
	Label      string        `gorm:"default:''"`
	Permission KeyPermission `gorm:"default:readwrite"`
	// Se true, a chave tem escopo global (todos os buckets do usuário).
	// Se false, acesso restrito às entradas em KeyBucketPermission.
	AllBuckets bool       `gorm:"default:true"`
	CreatedAt  *time.Time `gorm:"autoCreateTime"`

	BucketPerms []KeyBucketPermission `gorm:"foreignKey:KeyID;constraint:OnDelete:CASCADE"`
}

func (Key) TableName() string { return "key" }

// KeyBucketPermission define a permissão de uma chave para um bucket específico.
type KeyBucketPermission struct {
	ID         uint          `gorm:"primaryKey"`
	KeyID      uint          `gorm:"index;not null"`
	BucketID   string        `gorm:"index;not null"`
	Permission KeyPermission `gorm:"not null"`
}

func (KeyBucketPermission) TableName() string { return "key_bucket_permission" }
