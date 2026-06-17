import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { getRefreshToken, isAuthenticated } from '@/lib/auth'
import LoginPage from '@/pages/LoginPage'
import DashboardPage from '@/pages/DashboardPage'

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
