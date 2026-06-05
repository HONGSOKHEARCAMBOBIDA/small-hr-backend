package utils

import "github.com/gin-gonic/gin"

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("X-Frame-Options", "DENY")
		// ការពារ Website មិនឱ្យ Hacker យកទៅដាក់ក្នុង <iframe> ដើម្បីការពារអ្នកប្រើប្រាស់ពី Clickjacking Attack
		c.Header(
			"X-Content-Type-Options",
			"nosniff",
		)
		//  ប្រើសម្រាប់ បង្ខំ Browser ឱ្យគោរព Content-Type ដែល Server កំណត់ មិនឱ្យទស្សន៍ទាយខុសឡើយ។
		c.Header(
			"Content-Security-Policy",
			"default-src 'self'; font-src 'self' https://fonts.googleapis.com",
		)
		// Load បានតែអ្វីដែលមកពី Website ហាមទាំងអស់ដែលមកពីក្រៅ
		c.Header(
			"Referrer-Policy",
			"strict-origin",
		)
		// វា ប្រាប់តែ Domain មិនប្រាប់ Path និង Query ដែលមាន ព័ត៌មានសម្ងាត់៖
		c.Header(
			"Permissions-Policy",
			"geolocation=(), camera=(), microphone=()",
		)
		// ហាម Script ណាមួយក្នុង Website ប្រើ GPS, កាមេរ៉ា, មីក្រូហ្វូន របស់ User ទាំងស្រុង!
		c.Header(
			"Cross-Origin-Opener-Policy",
			"same-origin",
		)

		c.Header(
			"Strict-Transport-Security",
			"max-age=31536000; includeSubDomains",
		)

		//Browser អើយ! ១ឆ្នាំខាងមុខ បើអ្នកចូល Website នេះ ត្រូវប្រើ HTTPS តែប៉ុណ្ណោះ! HTTP ហាមដាច់ខាត!

		c.Next()
	}
}
