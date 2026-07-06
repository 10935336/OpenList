package op

import (
	"slices"

	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/pkg/errors"
)

func GetUserGroupById(id uint) (*model.UserGroup, error) {
	return db.GetUserGroupById(id)
}

func GetUserGroups(pageIndex, pageSize int) (groups []model.UserGroup, count int64, err error) {
	return db.GetUserGroups(pageIndex, pageSize)
}

func CreateUserGroup(g *model.UserGroup, memberIDs []uint) error {
	if err := db.CreateUserGroup(g); err != nil {
		return err
	}
	return syncGroupMembers(g.ID, memberIDs)
}

func UpdateUserGroup(g *model.UserGroup, memberIDs []uint) error {
	if _, err := db.GetUserGroupById(g.ID); err != nil {
		return err
	}
	if err := db.UpdateUserGroup(g); err != nil {
		return err
	}
	return syncGroupMembers(g.ID, memberIDs)
}

func DeleteUserGroupById(id uint) error {
	if _, err := db.GetUserGroupById(id); err != nil {
		return err
	}
	if err := syncGroupMembers(id, nil); err != nil {
		return errors.WithMessage(err, "failed to remove group from users")
	}
	return db.DeleteUserGroupById(id)
}

// syncGroupMembers makes the set of users whose GroupIDs contain groupID equal
// to memberIDs, going through UpdateUser so user caches are invalidated.
func syncGroupMembers(groupID uint, memberIDs []uint) error {
	users, _, err := db.GetUsers(1, model.MaxInt)
	if err != nil {
		return err
	}
	for i := range users {
		user := &users[i]
		has := slices.Contains(user.GroupIDs, groupID)
		want := slices.Contains(memberIDs, user.ID)
		if has == want {
			continue
		}
		if want {
			user.GroupIDs = append(user.GroupIDs, groupID)
		} else {
			user.GroupIDs = slices.DeleteFunc(user.GroupIDs, func(id uint) bool { return id == groupID })
		}
		if err := UpdateUser(user); err != nil {
			return errors.WithMessagef(err, "failed to update groups of user %s", user.Username)
		}
	}
	return nil
}
