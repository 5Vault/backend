package external

// emailTemplates maps template name → HTML body (Go text/template syntax).
var emailTemplates = map[string]string{
	"welcome":        tmplWelcome,
	"new_device":     tmplNewDevice,
	"password_reset": tmplPasswordReset,
	"tier_upgrade":   tmplTierUpgrade,
	"ticket_opened":  tmplTicketOpened,
	"ticket_reply":   tmplTicketReply,
	"ticket_closed":  tmplTicketClosed,
}

const emailBase = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>
  body{margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif}
  .wrap{max-width:560px;margin:40px auto;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08)}
  .header{background:#09090b;padding:28px 32px;display:flex;align-items:center;gap:10px}
  .logo{font-size:20px;font-weight:700;color:#fff;letter-spacing:-.5px}
  .logo span{color:#e23b3b}
  .body{padding:32px}
  h1{margin:0 0 8px;font-size:22px;color:#09090b;font-weight:700}
  p{margin:0 0 16px;font-size:15px;color:#52525b;line-height:1.6}
  .btn{display:inline-block;margin:8px 0 20px;padding:12px 28px;background:#e23b3b;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px}
  .code{display:inline-block;padding:10px 20px;background:#f4f4f5;border-radius:8px;font-family:monospace;font-size:18px;font-weight:700;color:#09090b;letter-spacing:2px;margin:8px 0 20px}
  .info{padding:16px;background:#f4f4f5;border-radius:8px;margin:16px 0}
  .info p{margin:0;font-size:14px;color:#71717a}
  .footer{padding:20px 32px;border-top:1px solid #f4f4f5}
  .footer p{margin:0;font-size:12px;color:#a1a1aa;text-align:center}
  .tag{display:inline-block;padding:3px 10px;background:#fef2f2;color:#e23b3b;border-radius:99px;font-size:12px;font-weight:600;margin-bottom:16px}
</style>
</head>
<body>
<div class="wrap">
  <div class="header">
    <div class="logo">Five<span>Vault</span></div>
  </div>
  <div class="body">
    {{.Content}}
  </div>
  <div class="footer">
    <p>© 2025 FiveKeepr · Você está recebendo este email pois tem uma conta na plataforma.</p>
  </div>
</div>
</body>
</html>`

const tmplWelcome = `<!DOCTYPE html>
<html lang="pt-BR"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif}.wrap{max-width:560px;margin:40px auto;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08)}.header{background:#09090b;padding:28px 32px}.logo{font-size:20px;font-weight:700;color:#fff;letter-spacing:-.5px}.logo span{color:#e23b3b}.body{padding:32px}h1{margin:0 0 8px;font-size:22px;color:#09090b;font-weight:700}p{margin:0 0 16px;font-size:15px;color:#52525b;line-height:1.6}.btn{display:inline-block;margin:8px 0 20px;padding:12px 28px;background:#e23b3b;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px}.footer{padding:20px 32px;border-top:1px solid #f4f4f5}.footer p{margin:0;font-size:12px;color:#a1a1aa;text-align:center}</style>
</head><body><div class="wrap"><div class="header"><div class="logo">Five<span>Vault</span></div></div>
<div class="body">
<h1>Bem-vindo, {{.Name}}! 👋</h1>
<p>Sua conta foi criada com sucesso. Você já pode fazer login e começar a usar o FiveKeepr.</p>
<p>Com o FiveKeepr você tem acesso a armazenamento seguro na nuvem, gerenciamento de buckets R2 e muito mais.</p>
<a href="{{.AppURL}}" class="btn">Acessar o FiveKeepr</a>
<p style="font-size:13px;color:#a1a1aa">Usuário: <strong>{{.Username}}</strong></p>
</div><div class="footer"><p>© 2025 FiveKeepr</p></div></div></body></html>`

const tmplNewDevice = `<!DOCTYPE html>
<html lang="pt-BR"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif}.wrap{max-width:560px;margin:40px auto;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08)}.header{background:#09090b;padding:28px 32px}.logo{font-size:20px;font-weight:700;color:#fff;letter-spacing:-.5px}.logo span{color:#e23b3b}.body{padding:32px}h1{margin:0 0 8px;font-size:22px;color:#09090b;font-weight:700}p{margin:0 0 16px;font-size:15px;color:#52525b;line-height:1.6}.btn{display:inline-block;margin:8px 0 20px;padding:12px 28px;background:#e23b3b;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px}.info{padding:16px;background:#fef2f2;border-radius:8px;margin:16px 0;border-left:3px solid #e23b3b}.info p{margin:0;font-size:14px;color:#71717a}.footer{padding:20px 32px;border-top:1px solid #f4f4f5}.footer p{margin:0;font-size:12px;color:#a1a1aa;text-align:center}</style>
</head><body><div class="wrap"><div class="header"><div class="logo">Five<span>Vault</span></div></div>
<div class="body">
<h1>🔐 Novo acesso detectado</h1>
<p>Detectamos um login na sua conta a partir de um novo endereço IP.</p>
<div class="info">
  <p><strong>IP:</strong> {{.IP}}</p>
  <p style="margin-top:6px"><strong>Horário:</strong> {{.Time}}</p>
</div>
<p>Se foi você, pode ignorar este email. Se não reconhece este acesso, recomendamos que redefina sua senha imediatamente.</p>
<a href="{{.ResetURL}}" class="btn">Redefinir minha senha</a>
</div><div class="footer"><p>© 2025 FiveKeepr</p></div></div></body></html>`

const tmplPasswordReset = `<!DOCTYPE html>
<html lang="pt-BR"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif}.wrap{max-width:560px;margin:40px auto;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08)}.header{background:#09090b;padding:28px 32px}.logo{font-size:20px;font-weight:700;color:#fff;letter-spacing:-.5px}.logo span{color:#e23b3b}.body{padding:32px}h1{margin:0 0 8px;font-size:22px;color:#09090b;font-weight:700}p{margin:0 0 16px;font-size:15px;color:#52525b;line-height:1.6}.btn{display:inline-block;margin:8px 0 20px;padding:12px 28px;background:#e23b3b;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px}.footer{padding:20px 32px;border-top:1px solid #f4f4f5}.footer p{margin:0;font-size:12px;color:#a1a1aa;text-align:center}</style>
</head><body><div class="wrap"><div class="header"><div class="logo">Five<span>Vault</span></div></div>
<div class="body">
<h1>Redefinição de senha</h1>
<p>Recebemos uma solicitação para redefinir a senha da sua conta FiveKeepr.</p>
<p>Clique no botão abaixo para criar uma nova senha. Este link é válido por <strong>1 hora</strong>.</p>
<a href="{{.ResetURL}}" class="btn">Redefinir senha</a>
<p style="font-size:13px;color:#a1a1aa">Se você não solicitou a redefinição, ignore este email. Sua senha permanece a mesma.</p>
</div><div class="footer"><p>© 2025 FiveKeepr</p></div></div></body></html>`

const tmplTierUpgrade = `<!DOCTYPE html>
<html lang="pt-BR"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif}.wrap{max-width:560px;margin:40px auto;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08)}.header{background:#09090b;padding:28px 32px}.logo{font-size:20px;font-weight:700;color:#fff;letter-spacing:-.5px}.logo span{color:#e23b3b}.body{padding:32px}h1{margin:0 0 8px;font-size:22px;color:#09090b;font-weight:700}p{margin:0 0 16px;font-size:15px;color:#52525b;line-height:1.6}.btn{display:inline-block;margin:8px 0 20px;padding:12px 28px;background:#e23b3b;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px}.badge{display:inline-block;padding:4px 14px;background:#fef2f2;color:#e23b3b;border-radius:99px;font-size:13px;font-weight:700;margin-bottom:16px}.info{padding:16px;background:#f4f4f5;border-radius:8px;margin:16px 0}.info p{margin:0;font-size:14px;color:#71717a}.footer{padding:20px 32px;border-top:1px solid #f4f4f5}.footer p{margin:0;font-size:12px;color:#a1a1aa;text-align:center}</style>
</head><body><div class="wrap"><div class="header"><div class="logo">Five<span>Vault</span></div></div>
<div class="body">
<div class="badge">{{.TierName}}</div>
<h1>Upgrade realizado! 🎉</h1>
<p>Seu plano foi atualizado com sucesso para <strong>{{.TierName}}</strong>.</p>
<div class="info">
  <p><strong>Plano:</strong> {{.TierName}}</p>
  <p style="margin-top:6px"><strong>Valor:</strong> R$ {{.Amount}}</p>
  <p style="margin-top:6px"><strong>Data:</strong> {{.Date}}</p>
</div>
<p>Agora você tem acesso a todos os recursos do plano {{.TierName}}. Aproveite!</p>
<a href="{{.AppURL}}" class="btn">Acessar o FiveKeepr</a>
</div><div class="footer"><p>© 2025 FiveKeepr</p></div></div></body></html>`

const tmplTicketOpened = `<!DOCTYPE html>
<html lang="pt-BR"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif}.wrap{max-width:560px;margin:40px auto;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08)}.header{background:#09090b;padding:28px 32px}.logo{font-size:20px;font-weight:700;color:#fff;letter-spacing:-.5px}.logo span{color:#e23b3b}.body{padding:32px}h1{margin:0 0 8px;font-size:22px;color:#09090b;font-weight:700}p{margin:0 0 16px;font-size:15px;color:#52525b;line-height:1.6}.btn{display:inline-block;margin:8px 0 20px;padding:12px 28px;background:#e23b3b;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px}.msg{padding:16px;background:#f4f4f5;border-radius:8px;margin:16px 0;border-left:3px solid #e23b3b}.footer{padding:20px 32px;border-top:1px solid #f4f4f5}.footer p{margin:0;font-size:12px;color:#a1a1aa;text-align:center}</style>
</head><body><div class="wrap"><div class="header"><div class="logo">Five<span>Vault</span></div></div>
<div class="body">
<h1>Novo ticket de suporte</h1>
<p><strong>{{.UserName}}</strong> ({{.UserEmail}}) abriu um novo ticket:</p>
<p><strong>Assunto:</strong> {{.Subject}}</p>
<div class="msg"><p>{{.FirstMessage}}</p></div>
<a href="{{.AdminURL}}" class="btn">Ver ticket no painel admin</a>
</div><div class="footer"><p>© 2025 FiveKeepr</p></div></div></body></html>`

const tmplTicketReply = `<!DOCTYPE html>
<html lang="pt-BR"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif}.wrap{max-width:560px;margin:40px auto;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08)}.header{background:#09090b;padding:28px 32px}.logo{font-size:20px;font-weight:700;color:#fff;letter-spacing:-.5px}.logo span{color:#e23b3b}.body{padding:32px}h1{margin:0 0 8px;font-size:22px;color:#09090b;font-weight:700}p{margin:0 0 16px;font-size:15px;color:#52525b;line-height:1.6}.btn{display:inline-block;margin:8px 0 20px;padding:12px 28px;background:#e23b3b;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px}.msg{padding:16px;background:#f4f4f5;border-radius:8px;margin:16px 0;border-left:3px solid #e23b3b}.footer{padding:20px 32px;border-top:1px solid #f4f4f5}.footer p{margin:0;font-size:12px;color:#a1a1aa;text-align:center}</style>
</head><body><div class="wrap"><div class="header"><div class="logo">Five<span>Vault</span></div></div>
<div class="body">
<h1>Resposta ao seu ticket</h1>
<p>O suporte FiveKeepr respondeu ao seu ticket: <strong>{{.Subject}}</strong></p>
<div class="msg"><p>{{.ReplyContent}}</p></div>
<a href="{{.TicketURL}}" class="btn">Ver ticket completo</a>
</div><div class="footer"><p>© 2025 FiveKeepr</p></div></div></body></html>`

const tmplTicketClosed = `<!DOCTYPE html>
<html lang="pt-BR"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif}.wrap{max-width:560px;margin:40px auto;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08)}.header{background:#09090b;padding:28px 32px}.logo{font-size:20px;font-weight:700;color:#fff;letter-spacing:-.5px}.logo span{color:#e23b3b}.body{padding:32px}h1{margin:0 0 8px;font-size:22px;color:#09090b;font-weight:700}p{margin:0 0 16px;font-size:15px;color:#52525b;line-height:1.6}.btn{display:inline-block;margin:8px 0 20px;padding:12px 28px;background:#e23b3b;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px}.footer{padding:20px 32px;border-top:1px solid #f4f4f5}.footer p{margin:0;font-size:12px;color:#a1a1aa;text-align:center}</style>
</head><body><div class="wrap"><div class="header"><div class="logo">Five<span>Vault</span></div></div>
<div class="body">
<h1>Ticket encerrado ✓</h1>
<p>Seu ticket <strong>{{.Subject}}</strong> foi marcado como resolvido.</p>
<p>Se o problema persistir ou você tiver novas dúvidas, abra um novo ticket a qualquer momento.</p>
<a href="{{.SupportURL}}" class="btn">Abrir novo ticket</a>
</div><div class="footer"><p>© 2025 FiveKeepr</p></div></div></body></html>`
