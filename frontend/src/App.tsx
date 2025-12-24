import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/layout/Layout';
import PerformanceOverviewPage from './pages/PerformanceOverview';
import SpecificCompetitionPage from './pages/SpecificCompetition';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Navigate to="/performance-overview" replace />} />
        <Route path="performance-overview" element={<PerformanceOverviewPage />} />
        <Route path="specific-competition" element={<SpecificCompetitionPage />} />
      </Route>
    </Routes>
  );
}

export default App;
