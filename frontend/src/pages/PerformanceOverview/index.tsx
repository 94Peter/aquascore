import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { PerformanceService, EventPerformance } from '../../api/generated';
import EventPerformanceCard from '../../components/performance/EventPerformanceCard';
import { useAthlete } from '../../context/AthleteContext';

const PerformanceOverviewPage: React.FC = () => {
  const { selectedAthlete } = useAthlete();

  const { data, isLoading, error } = useQuery({
    queryKey: ['performanceOverview', selectedAthlete],
    queryFn: () => {
      if (!selectedAthlete) {
        throw new Error("Athlete not selected");
      }
      return PerformanceService.getAthletesPerformanceOverview(selectedAthlete);
    },
    enabled: !!selectedAthlete,
  });

  if (!selectedAthlete) {
    return <h2 className="text-xl font-semibold text-slate-500">Please select an athlete from the dropdown above.</h2>;
  }

  if (isLoading) {
    return <h2 className="text-xl font-semibold text-slate-500">Loading performance overview for {selectedAthlete}...</h2>;
  }

  if (error) {
    return <h2 className="text-xl font-semibold text-red-500">An error has occurred: {error.message}</h2>;
  }

  return (
    <section>
      <h2 className="text-xl font-bold text-slate-900 mb-1">🏆 個人最佳成績總覽</h2>
      <p className="text-sm text-slate-500 mb-4">All-Time Personal Bests. 點擊卡片以深入分析項目表現。</p>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {data?.map((eventPerformance: EventPerformance) => (
          <EventPerformanceCard key={eventPerformance.event_name} eventPerformance={eventPerformance} />
        ))}
         {/* Placeholder for "add more" card if needed */}
         {/* <div className="bg-white rounded-xl shadow-sm border border-slate-200/80 p-5 flex flex-col justify-center items-center text-slate-400 border-dashed min-h-[260px]">
            <p className="mt-2 text-sm font-medium">更多項目...</p>
        </div> */}
      </div>
    </section>
  );
};

export default PerformanceOverviewPage;