package db

import (
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/pkg/errors"
)

func GetUserGroupById(id uint) (*model.UserGroup, error) {
	var g model.UserGroup
	if err := db.First(&g, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get user group")
	}
	return &g, nil
}

func CreateUserGroup(g *model.UserGroup) error {
	return errors.WithStack(db.Create(g).Error)
}

func UpdateUserGroup(g *model.UserGroup) error {
	return errors.WithStack(db.Save(g).Error)
}

func GetUserGroups(pageIndex, pageSize int) (groups []model.UserGroup, count int64, err error) {
	groupDB := db.Model(&model.UserGroup{})
	if err = groupDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get user groups count")
	}
	if err = groupDB.Order(columnName("id")).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&groups).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get find user groups")
	}
	return groups, count, nil
}

func DeleteUserGroupById(id uint) error {
	return errors.WithStack(db.Delete(&model.UserGroup{}, id).Error)
}
