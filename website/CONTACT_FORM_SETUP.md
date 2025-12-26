# Contact Form Setup (Formspree)

The contact form uses **Formspree** - a free service for handling form submissions without backend code.

## Setup Instructions

1. **Create Formspree Account**

   - Go to https://formspree.io
   - Sign up (free plan allows 50 submissions/month)

2. **Create New Form**

   - Click "New Form"
   - Enter form name: "Promenade Contact"
   - Copy your Form ID (looks like: `abc123xyz`)

3. **Update Contact Page**

   - Edit `website/content/contact.md`
   - Replace `YOUR_FORM_ID` with your actual Form ID:
     ```html
     <form action="https://formspree.io/f/YOUR_FORM_ID" method="POST"></form>
     ```

4. **Deploy**
   - Commit and push changes
   - GitHub Actions will auto-deploy

## Features Included

 Spam protection (reCAPTCHA)  
 Email notifications  
 Auto-reply to sender  
 File uploads (optional)  
 Custom thank you page

## Free Plan Limits

- 50 submissions/month
- 1000 submissions/month (paid plan $10)

## Alternative: EmailJS

If you prefer EmailJS:

1. Sign up at https://emailjs.com
2. Create email service
3. Update form to use EmailJS API
4. No backend needed

## Alternative: Custom API

For full control, create API endpoint:

```go
// internal/adapter/http/v1/handler/contact_handler.go
func (h *ContactHandler) Submit(c *gin.Context) {
    // Save to DB or send email
}
```
