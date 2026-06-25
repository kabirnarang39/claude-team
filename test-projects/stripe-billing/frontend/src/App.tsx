import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import BillingPage from './pages/billing/BillingPage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/billing" element={<BillingPage />} />
        <Route path="/" element={<Navigate to="/billing" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
