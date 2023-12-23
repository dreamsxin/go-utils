
## result

```go
result := db.Where("id = ?", id).Delete(&User{})
err = result.Error
num = result.RowsAffected
```

## Unscoped
移除软删除过滤
```go
func (r *UserRepository) DeleteSoftDeletedByID(id uint) error {
    result := r.db.Unscoped().Delete(&User{}, id)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return gorm.ErrRecordNotFound
    }
    return nil
}
```

## Hooks

- BeforeSave
- BeforeCreate
- AfterCreate
- AfterSave

```go
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
  u.UUID = uuid.New()

  if !u.IsValid() {
    err = errors.New("can't save invalid data")
  }
  return
}

func (u *User) AfterCreate(tx *gorm.DB) (err error) {
  if u.ID == 1 {
    tx.Model(u).Update("role", "admin")
  }
  return
}
```

## Serializer

- json
- gob
- unixtime

```go
type User struct {
  Name        []byte                 `gorm:"serializer:json"`
  Roles       Roles                  `gorm:"serializer:json"`
  Contracts   map[string]interface{} `gorm:"serializer:json"`
  JobInfo     Job                    `gorm:"type:bytes;serializer:gob"`
  CreatedTime int64                  `gorm:"serializer:unixtime;type:time"` // store int as datetime into database
}

type Roles []string

type Job struct {
  Title    string
  Location string
  IsIntern bool
}
```