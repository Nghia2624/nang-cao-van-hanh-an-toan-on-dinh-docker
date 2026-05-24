import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ErrorBoundary } from './components/ErrorBoundary'
import ChatWidget from './components/ChatWidget'
import Dashboard from './pages/Dashboard'
import Containers from './pages/Containers'
import ContainerDetail from './pages/ContainerDetail'
import LogsViewer from './pages/LogsViewer'
import Alerts from './pages/Alerts'
import AIInsights from './pages/AIInsights'

export default function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/containers" element={<Containers />} />
          <Route path="/containers/:id" element={<ContainerDetail />} />
          <Route path="/logs" element={<LogsViewer />} />
          <Route path="/alerts" element={<Alerts />} />
          <Route path="/alerts/:id" element={<Alerts />} />
          <Route path="/ai-insights" element={<AIInsights />} />
        </Routes>
        <ChatWidget />
      </BrowserRouter>
    </ErrorBoundary>
  )
}
