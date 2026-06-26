import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { getRefreshToken, isAuthenticated } from '@/lib/auth'
import LoginPage from '@/pages/LoginPage'
import DashboardPage from '@/pages/DashboardPage'
import FornecedoresPage from '@/pages/FornecedoresPage'
import CategoriasPage from '@/pages/CategoriasPage'
import ProdutosPage from '@/pages/ProdutosPage'
import AjustesEstoquePage from '@/pages/AjustesEstoquePage'
import ComprasPage from '@/pages/ComprasPage'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const hasSession = isAuthenticated() || getRefreshToken() !== null
  return hasSession ? <>{children}</> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
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
          path="/estoque"
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
          path="/*"
          element={
            <PrivateRoute>
              <DashboardPage />
            </PrivateRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}
