package db

import (
	"context"
	"fmt"
	"time"

	"ojbk.io/gopherCron/errors"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"github.com/sirupsen/logrus"

	"ojbk.io/gopherCron/common"
	"ojbk.io/gopherCron/utils"
)

const (
	// 用户表名
	UserTable = "user"
)

// createAdminUser 创建系统管理员
func createAdminUser() error {
	var (
		salt string
		err  error
	)

	salt = utils.RandomStr(6)
	if _, err = Database.Collection(UserTable).InsertOne(context.TODO(), &common.User{
		ID:         primitive.NewObjectID().Hex(),
		Account:    common.ADMIN_USER_ACCOUNT,
		Password:   utils.BuildPassword(common.ADMIN_USER_PASSWORD, salt),
		Salt:       salt,
		Name:       common.ADMIN_USER_NAME,
		Permission: common.ADMIN_USER_PERMISSION,
		CreateTime: time.Now().Unix(),
		Project:    []string{common.ADMIN_PROJECT},
	}); err != nil {
		logrus.WithField("Error", err).Error("goperCron create admin user error")
		return err
	}

	return nil
}

// CreateUser 创建新用户
// 这个操作只有admin账号可以操作
func CreateUser(user *common.User) error {
	var (
		err    error
		errObj errors.Error
	)
	if _, err = Database.Collection(UserTable).InsertOne(context.TODO(), &common.User{
		ID:         primitive.NewObjectID().Hex(),
		Account:    user.Account,
		Password:   user.Password,
		Salt:       user.Salt,
		Name:       user.Name,
		Permission: "user",
		CreateTime: time.Now().Unix(),
	}); err != nil {
		errObj = errors.ErrInternalError
		errObj.Log = "[User - GetUserWithAccount] CreateUser error:" + err.Error()
		return errObj
	}

	return nil
}

// GetUserWithAccount 通过账号获取用户信息
func GetUserWithAccount(account string) (*common.User, error) {
	var (
		res    *mongo.SingleResult
		user   *common.User
		errObj errors.Error
		err    error
	)
	res = Database.Collection(UserTable).FindOne(context.TODO(), bson.M{"account": account})
	if res.Err() != nil {
		errObj = errors.ErrInternalError
		errObj.Log = "[User - GetUserWithAccount] FindOne error:" + res.Err().Error()
		return nil, errObj
	}

	if err = res.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		errObj = errors.ErrInternalError
		errObj.Log = "[DB - GetUserWithAccount] get user with account error:" + err.Error()
		fmt.Println(errObj.Log)
		return nil, errObj
	}

	return user, nil
}

// GetUserInfo 获取用户信息
func GetUserInfo(uid string) (*common.User, error) {
	var (
		res    *mongo.SingleResult
		user   *common.User
		errObj errors.Error
		err    error
	)
	res = Database.Collection(UserTable).FindOne(context.TODO(), bson.M{"_id": uid})
	if res.Err() != nil {
		errObj = errors.ErrInternalError
		errObj.Log = "[User - GetUserWithAccount] FindOne error:" + res.Err().Error()
		return nil, errObj
	}

	if err = res.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		errObj = errors.ErrInternalError
		errObj.Log = "[DB - GetUserWithAccount] get user with account error:" + err.Error()
		fmt.Println(errObj.Log)
		return nil, errObj
	}

	return user, nil
}

// ChangePassword 修改用户密码
func ChangePassword(uid, password, salt string) error {
	var (
		res    *mongo.UpdateResult
		errObj errors.Error
		err    error
	)

	res, err = Database.Collection(UserTable).UpdateOne(context.TODO(),
		bson.M{"_id": uid},
		bson.M{"$set": bson.M{"password": password, "salt": salt}})

	if err != nil {
		errObj = errors.ErrInternalError
		errObj.Log = "[User - ChangePassword] UpdateOne error:" + err.Error()
		return errObj
	}

	if res.ModifiedCount != 1 {
		return errors.ErrDataNotFound
	}

	return nil
}
