package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

// EmailService structure pour l'envoi d'emails
type EmailService struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// NewEmailService crée une nouvelle instance du service email
func NewEmailService() *EmailService {
	return &EmailService{
		Host:     Env("EMAIL_HOST"),
		Port:     Env("EMAIL_PORT"),
		Username: Env("EMAIL_USERNAME"),
		Password: Env("EMAIL_PASSWORD"),
		From:     Env("EMAIL_FROM"),
	}
}

// SendPasswordResetEmail envoie un email de réinitialisation de mot de passe
func (es *EmailService) SendPasswordResetEmail(to, token, employeeName string) error {
	if es.Host == "" || es.Port == "" || es.Username == "" || es.Password == "" {
		return fmt.Errorf("configuration email incomplète")
	}

	resetURL := Env("RESET_URL") + token

	subject := "Réinitialisation de votre mot de passe - IPOS-STOCK"

	// Template HTML pour l'email
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Réinitialisation de mot de passe</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .button { display: inline-block; background: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .token-box { 
            background: #e8f4fd; 
            border: 2px solid #007bff; 
            border-radius: 8px; 
            padding: 20px; 
            margin: 20px 0; 
            text-align: center;
            font-family: 'Courier New', monospace;
        }
        .token-title { 
            font-weight: bold; 
            color: #007bff; 
            margin-bottom: 10px;
            font-size: 16px;
        }
        .token-value { 
            background: white; 
            border: 1px solid #ddd; 
            border-radius: 4px; 
            padding: 15px; 
            font-size: 18px; 
            font-weight: bold; 
            color: #333; 
            word-break: break-all;
            letter-spacing: 1px;
            margin: 10px 0;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .token-value:hover { background: #f8f9fa; }
        .copy-instruction { 
            font-size: 12px; 
            color: #666; 
            margin-top: 5px;
            font-style: italic;
        }
        .footer { background: #333; color: white; padding: 15px; text-align: center; font-size: 12px; }
        .warning { background: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; margin: 15px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Réinitialisation de mot de passe</h1>
        </div>
        <div class="content">
            <p>Bonjour {{.EmployeeName}},</p>
            <p>Nous avons reçu une demande de réinitialisation de mot de passe pour votre compte.</p> 
            
            <div class="token-box">
                <div class="token-title">🔑 Votre code de réinitialisation :</div>
                <div class="token-value" onclick="this.select()">{{.Token}}</div>
                <div class="copy-instruction">Cliquez sur le code pour le sélectionner et le copier</div>
            </div>
             
            <div class="warning">
                <strong>⚠️ Important :</strong>
                <ul>
                    <li>Ce Code expire dans 3 heures</li>
                    <li>Si vous n'avez pas demandé cette réinitialisation, ignorez cet email</li>
                    <li>Ne partagez jamais ce lien avec qui que ce soit</li>
                </ul>
            </div>
        </div>
        <div class="footer">
            <p>Cet email a été généré automatiquement, merci de ne pas y répondre.</p>
            <p>&copy; 2025 ICTECH - Tous droits réservés</p>
        </div>
    </div>
</body>
</html>`

	// Parse du template
	tmpl, err := template.New("passwordReset").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("erreur lors du parsing du template: %v", err)
	}

	// Données pour le template
	data := struct {
		EmployeeName string
		ResetURL     string
		Token        string
	}{
		EmployeeName: employeeName,
		ResetURL:     resetURL,
		Token:        token,
	}

	// Génération du contenu HTML
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("erreur lors de l'exécution du template: %v", err)
	}

	// Construction du message email
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", to, subject, body.String())

	// Configuration SMTP
	auth := smtp.PlainAuth("", es.Username, es.Password, es.Host)

	// Envoi de l'email
	err = smtp.SendMail(es.Host+":"+es.Port, auth, es.From, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("erreur lors de l'envoi de l'email: %v", err)
	}

	return nil
}

// SendVerificationEmail envoie un email de vérification de compte
func (es *EmailService) SendVerificationEmail(to, token, fullname string) error {
	if es.Host == "" || es.Port == "" || es.Username == "" || es.Password == "" {
		return fmt.Errorf("configuration email incomplète")
	}

	verificationURL := Env("FRONTEND_URL") + "/verify-email?token=" + token

	subject := "Vérification de votre compte - Ukarimu App"

	// Template HTML pour l'email de vérification
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Vérification de compte</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .button { display: inline-block; background: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .token-box { 
            background: #e8f4fd; 
            border: 2px solid #007bff; 
            border-radius: 8px; 
            padding: 20px; 
            margin: 20px 0; 
            text-align: center;
            font-family: 'Courier New', monospace;
        }
        .token-title { 
            font-weight: bold; 
            color: #007bff; 
            margin-bottom: 10px;
            font-size: 16px;
        }
        .token-value { 
            background: white; 
            border: 1px solid #ddd; 
            border-radius: 4px; 
            padding: 15px; 
            font-size: 18px; 
            font-weight: bold; 
            color: #333; 
            word-break: break-all;
            letter-spacing: 1px;
            margin: 10px 0;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .token-value:hover { background: #f8f9fa; }
        .copy-instruction { 
            font-size: 12px; 
            color: #666; 
            margin-top: 5px;
            font-style: italic;
        }
        .footer { background: #333; color: white; padding: 15px; text-align: center; font-size: 12px; }
        .warning { background: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; margin: 15px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎓 Bienvenue sur Ukarimu App</h1>
        </div>
        <div class="content">
            <p>Bonjour {{.Fullname}},</p>
            <p>Merci de vous être inscrit sur Ukarimu App ! Pour activer votre compte, veuillez vérifier votre adresse email.</p> 
            
            <div style="text-align: center;">
                <a href="{{.VerificationURL}}" class="button">Vérifier mon email</a>
            </div>

            <p>Ou copiez et collez le code de vérification ci-dessous dans l'application :</p>
            
            <div class="token-box">
                <div class="token-title">🔑 Votre code de vérification :</div>
                <div class="token-value" onclick="this.select()">{{.Token}}</div>
                <div class="copy-instruction">Cliquez sur le code pour le sélectionner et le copier</div>
            </div>
             
            <div class="warning">
                <strong>⚠️ Important :</strong>
                <ul>
                    <li>Ce code expire dans 24 heures</li>
                    <li>Si vous n'avez pas créé de compte, ignorez cet email</li>
                    <li>Ne partagez jamais ce code avec qui que ce soit</li>
                </ul>
            </div>
        </div>
        <div class="footer">
            <p>Cet email a été généré automatiquement, merci de ne pas y répondre.</p>
            <p>&copy; 2026 Ukarimu App - Tous droits réservés</p>
        </div>
    </div>
</body>
</html>`

	// Parse du template
	tmpl, err := template.New("emailVerification").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("erreur lors du parsing du template: %v", err)
	}

	// Données pour le template
	data := struct {
		Fullname        string
		VerificationURL string
		Token           string
	}{
		Fullname:        fullname,
		VerificationURL: verificationURL,
		Token:           token,
	}

	// Génération du contenu HTML
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("erreur lors de l'exécution du template: %v", err)
	}

	// Construction du message email
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", to, subject, body.String())

	// Configuration SMTP
	auth := smtp.PlainAuth("", es.Username, es.Password, es.Host)

	// Envoi de l'email
	err = smtp.SendMail(es.Host+":"+es.Port, auth, es.From, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("erreur lors de l'envoi de l'email: %v", err)
	}

	return nil
}
