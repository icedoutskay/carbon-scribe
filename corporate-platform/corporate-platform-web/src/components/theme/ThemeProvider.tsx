'use client'

import * as React from 'react'

interface ThemeProviderProps {
  children: React.ReactNode
  attribute?: string
  defaultTheme?: string
  enableSystem?: boolean
  disableTransitionOnChange?: boolean
}

export function ThemeProvider({ 
  children, 
  attribute = 'class',
  defaultTheme = 'light',
  enableSystem = true,
  disableTransitionOnChange = false 
}: ThemeProviderProps) {
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
    
    // Check for saved theme or system preference
    const savedTheme = localStorage.getItem('theme')
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    
    if (savedTheme === 'dark' || (!savedTheme && enableSystem && prefersDark)) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
    
    if (disableTransitionOnChange) {
      const css = document.createElement('style')
      css.textContent = `
        * {
          transition: none !important;
        }
      `
      document.head.appendChild(css)
      
      const forceReflow = () => document.body.offsetHeight
      forceReflow()
      
      setTimeout(() => {
        document.head.removeChild(css)
      }, 1)
    }
  }, [enableSystem, disableTransitionOnChange])

  const value = React.useMemo(() => ({
    theme: mounted ? (document.documentElement.classList.contains('dark') ? 'dark' : 'light') : defaultTheme,
    setTheme: (theme: string) => {
      localStorage.setItem('theme', theme)
      if (theme === 'dark') {
        document.documentElement.classList.add('dark')
      } else if (theme === 'light') {
        document.documentElement.classList.remove('dark')
      } else {
        // System theme
        localStorage.removeItem('theme')
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
        if (prefersDark) {
          document.documentElement.classList.add('dark')
        } else {
          document.documentElement.classList.remove('dark')
        }
      }
    },
  }), [mounted, defaultTheme])

  if (!mounted) {
    return <>{children}</>
  }

  return (
    <div data-theme={value.theme} className={attribute === 'class' ? value.theme : ''}>
      {children}
    </div>
  )
}
