import { useAthlete } from '../../context/AthleteContext';
import { useQuery } from '@tanstack/react-query';
import { PerformanceService, ResultComparison, CompetitorComparison } from '../../api/generated';
import {
  ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, ReferenceLine, Label
} from 'recharts';
import { X, LoaderCircle, Info, ChevronsUp, ChevronDown } from 'lucide-react';
import clsx from 'clsx';
import CustomTooltip from '../common/Tooltip'; // Assuming the Tooltip component file is named Tooltip.tsx
import { formatTime } from '../../utils/timeFormatter';

interface RaceComparisonModalProps {
  raceId: string;
  open: boolean;
  onClose: () => void;
}

const getDiffClasses = (diffLabel?: CompetitorComparison.diff_label) => {
  switch (diffLabel) {
    case 'far_ahead':
    case 'slightly_ahead':
      return 'text-green-600';
    case 'far_behind':
    case 'slightly_behind':
      return 'text-red-600';
    case 'your_result':
      return 'text-slate-700 font-bold';
    default:
      return 'text-slate-500';
  }
};

const getDiffIcon = (diffLabel?: CompetitorComparison.diff_label) => {
    switch (diffLabel) {
        case 'far_ahead':
        case 'slightly_ahead':
          return <ChevronsUp className="w-4 h-4 inline-block" />;
        case 'far_behind':
        case 'slightly_behind':
          return <ChevronDown className="w-4 h-4 inline-block" />;
        default:
          return null;
      }
}

const RaceComparisonModal: React.FC<RaceComparisonModalProps> = ({ raceId, open, onClose }) => {
  const { selectedAthlete } = useAthlete(); // Get selectedAthlete from context

  const { data: comparisonData, isLoading, error } = useQuery<ResultComparison, Error>({
    queryKey: ['raceComparison', raceId, selectedAthlete], // Add selectedAthlete to queryKey
    queryFn: () => PerformanceService.getRaceComparison(raceId, selectedAthlete!), // Pass athleteName
    enabled: open && !!selectedAthlete, // Enable only if modal is open and athlete is selected
  });

  if (!open) {
    return null;
  }

  // If no athlete is selected, and the modal somehow opened, or if it opens before athlete selection
  if (!selectedAthlete) {
    return (
      <div
        className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex justify-center items-center p-4 transition-opacity duration-300"
        onClick={onClose}
      >
        <div
          className="bg-slate-50 rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] flex flex-col"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex-shrink-0 p-5 border-b border-slate-200 flex justify-between items-center">
            <h2 className="text-lg font-bold text-slate-900">錯誤</h2>
            <button onClick={onClose} className="p-1 rounded-full text-slate-500 hover:bg-slate-200 hover:text-slate-800 transition">
              <X className="w-5 h-5" />
            </button>
          </div>
          <div className="flex-grow p-6 text-center text-red-500">
            請選擇一位運動員以查看比賽比較。
          </div>
        </div>
      </div>
    );
  }

  const yourRank = comparisonData?.target_result?.rank;

  return (
    <div
      className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex justify-center items-center p-4 transition-opacity duration-300"
      onClick={onClose}
    >
      <div
        className="bg-slate-50 rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Modal Header */}
        <div className="flex-shrink-0 p-5 border-b border-slate-200 flex justify-between items-center">
          <div>
            <h2 className="text-lg font-bold text-slate-900">對比分析: {comparisonData?.target_result?.event_name}</h2>
            <p className="text-sm text-slate-500">{comparisonData?.target_result?.competition_name}</p>
          </div>
          <button onClick={onClose} className="p-1 rounded-full text-slate-500 hover:bg-slate-200 hover:text-slate-800 transition">
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Content */}
        <div className="flex-grow p-6 overflow-y-auto space-y-8">
          {isLoading && (
            <div className="flex justify-center items-center h-96">
              <LoaderCircle className="w-12 h-12 text-slate-400 animate-spin" />
            </div>
          )}
          {error && <div className="text-red-500">Error: {error.message}</div>}
          {comparisonData && (
            <>
              {/* Charts Section */}
              <div>
                <h3 className="font-semibold text-slate-800 mb-1 flex items-center gap-1.5">
                  📈 成績分佈圖 & 差距柱狀圖
                  <CustomTooltip text="快速定位您在賽場上的競爭位置。分佈圖看清全貌，差距圖聚焦於您與他人的秒數差距。">
                    <Info className="w-4 h-4 text-slate-400" />
                  </CustomTooltip>
                </h3>
                <div className="mt-4 p-4 h-64 bg-white rounded-lg border border-slate-200/80">
                    <ResponsiveContainer width="100%" height="100%">
                        <ScatterChart margin={{ top: 20, right: 30, bottom: 20, left: 20 }}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis type="number" dataKey="time" name="Time" unit="s" domain={['dataMin - 0.5', 'dataMax + 0.5']} tickFormatter={(value: number) => formatTime(value)} />
                            <YAxis type="category" dataKey="name" name="Athlete" hide={true} />
                            <Tooltip cursor={{ strokeDasharray: '3 3' }} formatter={(value: unknown, _name: string | undefined, props: { payload?: { name: string } }) => [`${props.payload?.name || 'N/A'}: ${formatTime(value as number)}`, null]} />
                            <Legend />
                            <ReferenceLine x={comparisonData.records?.national_record?.time} stroke="purple" strokeDasharray="4 4">
                                <Label value="全國REC" position="top" fill="purple" fontSize={12}/>
                            </ReferenceLine>
                             <ReferenceLine x={comparisonData.records?.games_record?.time} stroke="gray" strokeDasharray="4 4">
                                <Label value="大會REC" position="top" fill="gray" fontSize={12} />
                            </ReferenceLine>
                            <Scatter name="Competitors" data={comparisonData.competitor_comparison?.map(c => ({ time: c.record_time, name: c.athlete_name, rank: c.rank }))} fill="#8884d8">
                            </Scatter>
                        </ScatterChart>
                    </ResponsiveContainer>
                </div>
              </div>

              {/* Detailed Data Section */}
              <div>
                <h3 className="font-semibold text-slate-800 mb-4 flex items-center gap-1.5">📋 詳細數據</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6 text-center">
                  {/* Your Result */}
                  <div className="bg-white p-3 rounded-lg border border-blue-300 shadow-sm">
                    <p className="text-sm text-slate-500">您的成績 (第 {yourRank} 名)</p>
                    <p className="text-xl font-bold text-blue-600">{formatTime(comparisonData.target_result?.record_time)}</p>
                  </div>
                  {/* Games Record */}
                  <div className="bg-white p-3 rounded-lg border border-slate-200/80">
                    <p className="text-sm text-slate-500 flex items-center justify-center gap-1">🏆 大會紀錄</p>
                    <p className="text-xl font-bold text-slate-800">
                      {formatTime(comparisonData.records?.games_record?.time)}
                      <span className="text-sm font-medium text-green-600 ml-1">({comparisonData.records?.games_record?.diff})</span>
                    </p>
                  </div>
                  {/* National Record */}
                  <div className="bg-white p-3 rounded-lg border border-slate-200/80">
                    <p className="text-sm text-slate-500 flex items-center justify-center gap-1">🌟 全國紀錄</p>
                    <p className="text-xl font-bold text-purple-800">
                      {formatTime(comparisonData.records?.national_record?.time)}
                      <span className="text-sm font-medium text-green-600 ml-1">({comparisonData.records?.national_record?.diff})</span>
                    </p>
                  </div>
                </div>
                
                <div className="flow-root">
                  <div className="-mx-6 -my-2 overflow-x-auto">
                    <div className="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
                      <table className="min-w-full divide-y divide-slate-200">
                        <thead className="bg-slate-100">
                          <tr>
                            <th scope="col" className="py-3.5 pl-6 pr-3 text-left text-sm font-semibold text-slate-900">名次</th>
                            <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900">姓名</th>
                            <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900">成績</th>
                            <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 flex items-center gap-1">與您差距</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-200 bg-white">
                          {comparisonData.competitor_comparison?.map((comp) => (
                            <tr key={comp.rank} className={clsx(comp.diff_label === 'your_result' && 'bg-blue-50')}>
                              <td className="whitespace-nowrap py-4 pl-6 pr-3 text-sm font-medium">
                                <span className={clsx(comp.diff_label === 'your_result' ? 'text-blue-700' : 'text-slate-800')}>{comp.rank}</span>
                              </td>
                              <td className="whitespace-nowrap px-3 py-4 text-sm">
                                <span className={clsx(comp.diff_label === 'your_result' ? 'text-blue-700 font-semibold' : 'text-slate-500')}>{comp.athlete_name}</span>
                              </td>
                              <td className="whitespace-nowrap px-3 py-4 text-sm font-semibold text-slate-800">{formatTime(comp.record_time)}</td>
                              <td className={clsx("whitespace-nowrap px-3 py-4 text-sm font-semibold", getDiffClasses(comp.diff_label))}>
                                <span className="inline-flex items-center gap-1.5">
                                  {comp.diff_from_target} {getDiffIcon(comp.diff_label)}
                                </span>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default RaceComparisonModal;

