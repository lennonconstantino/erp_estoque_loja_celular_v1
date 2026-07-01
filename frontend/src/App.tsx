import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { getRefreshToken, isAuthenticated } from '@/lib/auth'
import { ThemeProvider } from '@/lib/theme'
import { Toaster } from '@/components/ui/sonner'
import LoginPage from '@/pages/LoginPage'
import DashboardPage from '@/pages/DashboardPage'
import FornecedoresPage from '@/pages/FornecedoresPage'
import CategoriasPage from '@/pages/CategoriasPage'
import ProdutosPage from '@/pages/ProdutosPage'
import ClientesPage from '@/pages/ClientesPage'
import AjustesEstoquePage from '@/pages/AjustesEstoquePage'
import ComprasPage from '@/pages/ComprasPage'
import VendasPage from '@/pages/VendasPage'
import NovaVendaPage from '@/pages/NovaVendaPage'
import RelatoriosPage from '@/pages/RelatoriosPage'
import UsuariosPage from '@/pages/UsuariosPage'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const hasSession = isAuthenticated() || getRefreshToken() !== null
  return hasSession ? <>{children}</> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
        <Route
          path="/clientes"
          element={
            <PrivateRoute>
              <ClientesPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/fornecedores"
          element={
            <PrivateRoute>
              <FornecedoresPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/categorias"
          element={
            <PrivateRoute>
              <CategoriasPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/produtos"
          element={
            <PrivateRoute>
              <ProdutosPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/estoque/ajustes"
          element={
            <PrivateRoute>
              <AjustesEstoquePage />
            </PrivateRoute>
          }
        />
        <Route
          path="/compras"
          element={
            <PrivateRoute>
              <ComprasPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/vendas"
          element={
            <PrivateRoute>
              <VendasPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/vendas/nova"
          element={
            <PrivateRoute>
              <NovaVendaPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/relatorios"
          element={
            <PrivateRoute>
              <RelatoriosPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/usuarios"
          element={
            <PrivateRoute>
              <UsuariosPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/*"
          element={
            <PrivateRoute>
              <DashboardPage />
            </PrivateRoute>
          }
        />
        </Routes>
      </BrowserRouter>
      <Toaster position="top-right" richColors />
    </ThemeProvider>
  )
}
