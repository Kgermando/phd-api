package database

import (
	"fmt"
	"log"

	"github.com/kgermando/phd-api/models"
	"github.com/kgermando/phd-api/utils"
)

// SeedSuperAdmin crée un user SuperAdmin uniquement si la table users est vide.
func SeedSuperAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		fmt.Println("SuperAdmin déjà existant, seed ignoré.")
		return
	}

	superAdmin := models.User{
		UUID:      utils.GenerateUUID(),
		Fullname:  "Support IT",
		Email:     "superadmin@dentic.app",
		Telephone: "+000000000000",
		Role:      "SuperAdmin",
		Permission      : "all",
		Status:    true,
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