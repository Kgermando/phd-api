package database

import (
	"fmt"
	"log"

	"github.com/kgermando/phd-api/models"
	"github.com/kgermando/phd-api/utils"
)

// SeedSuperAdmin crée un user SuperAdmin uniquement s'il n'en existe pas déjà un.
func SeedSuperAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", "Admin").Count(&count)
	if count > 0 {
		fmt.Println("Admin déjà existant, seed ignoré.")
		return
	}

	superAdmin := models.User{
		UUID:       utils.GenerateUUID(),
		Fullname:   "Support IT",
		Email:      "suport@phdc.app",
		Telephone:  "+000000000000",
		Role:       "Admin",
		Permission: "all",
		Status:     true,
	}

	rawPassword := "SuperAdmin@2026!"
	superAdmin.SetPassword(rawPassword)

	if err := DB.Create(&superAdmin).Error; err != nil {
		log.Printf("Échec de la création du SuperAdmin : %v\n", err)
		return
	}

	fmt.Println("SuperAdmin créé avec succès ✅")
	fmt.Printf("  Email    : %s\n", superAdmin.Email)
	fmt.Printf("  Password : %s\n", rawPassword)
	fmt.Println("  ⚠️  Changez ce mot de passe après la première connexion !")
}
