import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'MailTools - 发送邮件',
  description: '简洁的在线邮件发送工具',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN">
      <body className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
        {children}
      </body>
    </html>
  )
}
