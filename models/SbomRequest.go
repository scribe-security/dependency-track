package models

import (
	"deptrack/client"
	"fmt"

	"gorm.io/gorm"
)

var DB *gorm.DB

type SbomRequest struct {
	gorm.Model
	Sbom_raw string
	client.DepTrackSbomPostResponse
	Status string
}

func (p *SbomRequest) BeforeCreate(db *gorm.DB) error {
	fmt.Println("Before Create")
	return nil
}

func (p *SbomRequest) BeforeUpdate(db *gorm.DB) error {
	fmt.Println("Before Update")
	return nil
}

//create a user
func CreateSbomRequest(r *SbomRequest) (err error) {
	err = DB.Create(r).Error
	if err != nil {
		return err
	}
	fmt.Printf("Err: %+v\n", err)
	return nil
}

//get users
func GetSbomRequests(r *[]SbomRequest) (err error) {
	err = DB.Find(r).Error
	if err != nil {
		return err
	}
	return nil
}

//get user by id
func GetSbomRequest(r *SbomRequest, id uint) (err error) {
	err = DB.Where("id = ?", id).First(r).Error
	if err != nil {
		return err
	}
	return nil
}

func GetSbomRequestByStatus(r *SbomRequest, status string) (err error) {
	err = DB.Where("status = ?", status).First(r).Error
	if err != nil {
		return err
	}
	return nil
}

//update user
func UpdateSbomRequest(r *SbomRequest) (err error) {
	DB.Save(r)
	return nil
}

//delete user
func DeleteSbomRequest(r *SbomRequest, id string) (err error) {
	DB.Where("id = ?", id).Delete(r)
	return nil
}
