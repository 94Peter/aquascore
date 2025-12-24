import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { DataRetrievalService, Competition, AthleteRaceResult } from '../../api/generated';
import { useAthlete } from '../../context/AthleteContext';
import RaceComparisonModal from '../../components/performance/RaceComparisonModal';
import { Info, LoaderCircle } from 'lucide-react';
import { formatTime } from '../../utils/timeFormatter';

const SpecificCompetitionPage: React.FC = () => {
  const { selectedAthlete } = useAthlete();
  const [selectedYear, setSelectedYear] = useState('');
  const [selectedCompetition, setSelectedCompetition] = useState('');
  const [selectedRaceIdForComparison, setSelectedRaceIdForComparison] = useState<string | null>(null);

  const { data: years, isLoading: isLoadingYears } = useQuery({
    queryKey: ['years'],
    queryFn: () => DataRetrievalService.getYears(),
  });

  const { data: competitions, isLoading: isLoadingCompetitions } = useQuery({
    queryKey: ['competitions', selectedYear, selectedAthlete],
    queryFn: () => DataRetrievalService.getCompetitions(selectedYear, selectedAthlete || undefined),
    enabled: !!selectedYear && !!selectedAthlete,
  });

  const { data: races, isLoading: isLoadingRaces } = useQuery({
    queryKey: ['races', selectedYear, selectedCompetition, selectedAthlete],
    queryFn: () => {
      if (!selectedAthlete) {
        throw new Error("Athlete not selected");
      }
      return DataRetrievalService.getAthletesRaces(
        selectedAthlete,
        selectedCompetition,
        selectedYear,
      );
    },
    enabled: !!selectedYear && !!selectedCompetition && !!selectedAthlete,
  });

  if (!selectedAthlete) {
    return <h2 className="text-xl font-semibold text-slate-500">Please select an athlete from the dropdown above.</h2>;
  }

  const handleOpenComparison = (raceId: string) => {
    setSelectedRaceIdForComparison(raceId);
  };

  const handleCloseComparison = () => {
    setSelectedRaceIdForComparison(null);
  };

  return (
    <section>
      <h2 className="text-xl font-bold text-slate-900 mb-1">🏊 特定競賽查詢</h2>
      <p className="text-sm text-slate-500 mb-4">查詢在特定比賽中的所有項目成績，並與其他選手進行對比。</p>
      
      <div className="bg-white rounded-xl shadow-sm border border-slate-200/80 p-5">
        <div className="flex flex-col md:flex-row items-start md:items-center gap-2 md:gap-6">
          <h3 className="font-semibold text-slate-800 flex-shrink-0">篩選條件:</h3>
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 text-sm w-full md:w-auto">
            <div className="w-full sm:w-auto">
              <label htmlFor="comp-year" className="font-medium text-slate-600 block mb-1">年份</label>
              <select
                id="comp-year"
                value={selectedYear}
                onChange={(e) => {
                  setSelectedYear(e.target.value);
                  setSelectedCompetition(''); // Reset competition when year changes
                }}
                disabled={isLoadingYears}
                className="w-full bg-white border border-slate-300 rounded-md py-2 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="" disabled>Select Year</option>
                {years?.map((year: string) => <option key={year} value={year}>{year}</option>)}
              </select>
            </div>
            <div className="w-full sm:w-auto">
              <label htmlFor="comp-name" className="font-medium text-slate-600 block mb-1">競賽名稱</label>
              <select
                id="comp-name"
                value={selectedCompetition}
                onChange={(e) => setSelectedCompetition(e.target.value)}
                disabled={!selectedYear || isLoadingCompetitions}
                className="w-full bg-white border border-slate-300 rounded-md py-2 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="" disabled>Select Competition</option>
                {competitions?.filter(c => c.name).map((comp: Competition) => (
                  <option key={comp.name} value={comp.name!}>
                    {comp.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
        <p className="mt-3 text-xs text-slate-500 flex items-center gap-1.5">
          <Info className="w-3.5 h-3.5 flex-shrink-0" />
          <span>請先選擇年份，以動態載入該年度的競賽列表。</span>
        </p>

        <div className="mt-6 flow-root">
          <div className="-mx-5 -my-2 overflow-x-auto">
            <div className="inline-block min-w-full py-2 align-middle">
              <table className="min-w-full divide-y divide-slate-200">
                <thead className="bg-slate-50">
                  <tr>
                    <th scope="col" className="py-3.5 pl-5 pr-3 text-left text-sm font-semibold text-slate-900">項目</th>
                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900">個人成績</th>
                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900">名次</th>
                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900">備註</th>
                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900">成績對比</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200 bg-white">
                  {isLoadingRaces ? (
                    <tr>
                      <td colSpan={5} className="text-center py-10">
                        <LoaderCircle className="w-8 h-8 text-slate-400 animate-spin inline-block" />
                      </td>
                    </tr>
                  ) : races && races.length > 0 ? (
                    races.map((race: AthleteRaceResult) => (
                      <tr key={race.race_id}>
                        <td className="whitespace-nowrap py-4 pl-5 pr-3 text-sm font-semibold text-slate-600">{race.event_name}</td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-800 font-bold">{formatTime(race.record)}</td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-800 font-bold">{race.rank}</td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-500">{race.note}</td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm">
                          <button
                            onClick={() => handleOpenComparison(race.race_id!)}
                            className="font-semibold text-blue-600 hover:text-blue-800"
                          >
                            查看
                          </button>
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={5} className="text-center py-10 text-slate-500">
                        No races found for the selected criteria.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      {selectedRaceIdForComparison && (
        <RaceComparisonModal
          raceId={selectedRaceIdForComparison}
          open={!!selectedRaceIdForComparison}
          onClose={handleCloseComparison}
        />
      )}
    </section>
  );
};

export default SpecificCompetitionPage;