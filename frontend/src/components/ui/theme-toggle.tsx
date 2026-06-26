import { Moon, Sun } from 'lucide-react'
import { useTheme } from '@/lib/theme'
import { Button } from './button'

export function ThemeToggle() {
  const { theme, toggleTheme } = useTheme()

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleTheme}
      title={theme === 'light' ? 'Mudar para modo escuro' : 'Mudar para modo claro'}
    >
      {theme === 'light' ? <Moon className="w-4 h-4" /> : <Sun className="w-4 h-4" />}
    </Button>
  )
}
