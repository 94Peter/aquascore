import React from 'react';
import { Outlet, useLocation, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useAthlete } from '../../context/AthleteContext';
import { DataRetrievalService } from '../../api/generated';
import AthleteSelector from '../common/AthleteSelector';
import { Waves } from 'lucide-react';
import clsx from 'clsx';

const Layout: React.FC = () => {
  const location = useLocation();
  const { selectedAthlete, setSelectedAthlete } = useAthlete();

  const { data: athletes, isLoading } = useQuery({
    queryKey: ['allAthletes'],
    queryFn: () => DataRetrievalService.getAthletes(),
  });

  const tabs = [
    { name: '整體表現分析', path: '/performance-overview' },
    { name: '特定競賽查詢', path: '/specific-competition' },
  ];

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      <header className="flex flex-col md:flex-row justify-between items-center pb-6 border-b border-slate-200">
        <div className="flex items-center gap-3 mb-4 md:mb-0">
          <div className="w-10 h-10 bg-blue-600/10 text-blue-600 flex items-center justify-center rounded-lg">
            <Waves size={24} />
          </div>
          <h1 className="text-2xl font-bold text-slate-900">AquaScore 儀表板</h1>
        </div>
        <div className="w-full md:w-auto min-h-[42px]">
          <AthleteSelector
            onSelectAthlete={setSelectedAthlete}
            selectedAthlete={selectedAthlete}
            athletes={athletes}
            isLoading={isLoading}
          />
        </div>
      </header>

      <main className="mt-6">
        <div className="border-b border-slate-200">
          <nav className="-mb-px flex space-x-6" aria-label="Tabs">
            {tabs.map((tab) => (
              <Link
                key={tab.name}
                to={tab.path}
                className={clsx(
                  'whitespace-nowrap py-3 px-1 border-b-2 font-semibold text-sm',
                  location.pathname === tab.path
                    ? 'text-blue-600 border-blue-600'
                    : 'text-slate-500 border-transparent hover:text-slate-700 hover:border-slate-300'
                )}
              >
                {tab.name}
              </Link>
            ))}
          </nav>
        </div>
        <div className="mt-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
};

export default Layout;