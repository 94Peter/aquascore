import React, { useState } from 'react';
import { Analysis, EventPerformance, StabilityMetric, TrendMetric } from '../../api/generated';
import { ArrowRight, Info, Star, TrendingUp, Waves, ChevronDown, ChevronsUp } from 'lucide-react';
import clsx from 'clsx';
import { LineChart, Line, ResponsiveContainer, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ReferenceLine, BarChart, Bar } from 'recharts';
import Tooltip from '../common/Tooltip';
import { formatTime } from '../../utils/timeFormatter';

interface EventPerformanceCardProps {
  eventPerformance: EventPerformance;
}

// Helper functions (as before)
const getPbFreshnessClasses = (label?: Analysis.label) => {
  switch (label) {
    case Analysis.label.HOT_STREAK:
      return { bg: 'bg-green-500', text: 'text-green-600', label: 'Hot Streak' };
    case Analysis.label.STABLE:
      return { bg: 'bg-yellow-500', text: 'text-yellow-600', label: 'Stable' };
    case Analysis.label.NOT_UPDATED_RECENTLY:
      return { bg: 'bg-red-500', text: 'text-red-600', label: '久未更新' };
    default:
      return { bg: 'bg-slate-500', text: 'text-slate-600', label: 'N/A' };
  }
};

const getStabilityClasses = (label?: StabilityMetric.label) => {
  switch (label) {
    case StabilityMetric.label.HIGH:
      return { text: 'text-green-600', label: '🟢 高' };
    case StabilityMetric.label.MEDIUM:
      return { text: 'text-yellow-600', label: '🟡 中等' };
    case StabilityMetric.label.LOW:
      return { text: 'text-red-600', label: '🔴 低' };
    default:
      return { text: 'text-slate-600', label: 'N/A' };
  }
};

const getTrendIcon = (label?: TrendMetric.label) => {
  switch (label) {
    case TrendMetric.label.IMPROVING:
      return { icon: <ChevronsUp className="w-4 h-4" />, label: '進步中' };
    case TrendMetric.label.STABLE:
      return { icon: <ArrowRight className="w-4 h-4" />, label: '穩定' };
    case TrendMetric.label.DECLINING:
      return { icon: <ChevronDown className="w-4 h-4" />, label: '退步中' };
    default:
      return { icon: null, label: 'N/A' };
  }
};


const EventPerformanceCard: React.FC<EventPerformanceCardProps> = ({ eventPerformance }) => {
  const { event_name, personal_best, analysis, charts, recent_races } = eventPerformance;
  const [isExpanded, setIsExpanded] = useState(false);

  // Memoize or compute data for charts and insights
  const pbFreshness = getPbFreshnessClasses(analysis?.pb_freshness?.label);
  const stability = getStabilityClasses(analysis?.stability?.label);
  const trend = getTrendIcon(analysis?.trend?.label);
  
  const sparklineData = charts?.sparkline?.map(time => ({ time: -time }));
  
  const trendChartData = charts?.trend_chart?.dates?.map((date, index) => ({
    date: new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    time: -(charts?.trend_chart?.times?.[index] || 0), // Convert to negative for inverted axis trick
  }));
  
  // Calculate min and max for Y-axis domain from negative values
  const times = trendChartData?.map(dataPoint => dataPoint.time).filter((time): time is number => time !== undefined);
  const minTime = times && times.length > 0 ? Math.min(...times) : undefined; // min will be more negative (original max)
  const maxTime = times && times.length > 0 ? Math.max(...times) : undefined; // max will be less negative (original min)

  // Add some padding to the domain (for negative values)
  const yAxisDomain = (minTime !== undefined && maxTime !== undefined)
    ? [minTime * 1.05, maxTime * 0.95] // Apply padding. minTime is most negative, maxTime is least negative
    : undefined;
  
  const recentRacesVsPb = recent_races?.slice(0, 3).map((race, index) => ({
      name: `-${2-index}場`,
      diff: race.time && personal_best?.time ? race.time - personal_best.time : 0,
  })).reverse();

  const lastRace = recent_races?.[0];
  const lastRaceDiff = lastRace?.time && personal_best?.time ? lastRace.time - personal_best.time : 0;
  
  const recentThreeAvg = recent_races && recent_races.length >= 3 
    ? recent_races.slice(0, 3).reduce((acc, race) => acc + (race.time || 0), 0) / 3
    : 0;
  const recentThreeAvgDiff = recentThreeAvg && personal_best?.time ? recentThreeAvg - personal_best.time : 0;


  return (
    <div className={clsx("bg-white rounded-xl shadow-sm border p-5 transition-all duration-300", isExpanded ? "lg:col-span-2 border-blue-500 shadow-md" : "hover:border-blue-500 hover:shadow-md")}>
      {/* Condensed View */}
      <div className={clsx("flex justify-between items-start", isExpanded && "mb-6")}>
        <h3 className="text-lg font-semibold text-slate-800 flex items-center gap-2">
            <Waves className="w-5 h-5 text-blue-500" /> {event_name}
        </h3>
        <button
          onClick={(e) => { e.stopPropagation(); setIsExpanded(!isExpanded); }}
          className="text-sm font-semibold text-blue-600 hover:text-blue-800 flex items-center gap-1"
        >
          {isExpanded ? '收起分析' : '深入分析'} <ArrowRight className={clsx("w-4 h-4 transition-transform", isExpanded && "rotate-90")} />
        </button>
      </div>

       {/* Expanded View */}
      {isExpanded && (
          <div className="border-t border-slate-200 mt-6 pt-6 animate-fade-in">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-4">
              <h2 className="text-lg font-bold text-slate-900 mb-2 md:mb-0">🎯 項目洞察分析</h2>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
              {/* Left Panel: Charts */}
              <div className="lg:col-span-3 bg-slate-50/70 rounded-xl p-4">
                <h3 className="font-semibold text-slate-800">📈 成績趨勢圖</h3>
                <p className="text-xs text-slate-500">過去一年表現</p>
                <div className="mt-4 h-56">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={trendChartData} margin={{ top: 5, right: 20, left: -10, bottom: 5 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                      <XAxis dataKey="date" fontSize={12} tickLine={false} axisLine={false} />
                      <YAxis fontSize={12} tickLine={false} axisLine={false} domain={yAxisDomain} tickFormatter={(time) => formatTime(Math.abs(time as number), 'N/A')} />
                      <RechartsTooltip formatter={(value: unknown) => typeof value === 'number' ? [formatTime(Math.abs(value as number), 'N/A'), 'Time'] : [String(value), 'Time']} />
                      <Legend />
                      <Line type="monotone" dataKey="time" stroke="#3b82f6" strokeWidth={2} dot={{ r: 4 }} activeDot={{ r: 6 }} name="成績" />
                      {charts?.trend_chart?.pb_line && <ReferenceLine y={charts.trend_chart.pb_line} label={{ value: `PB: ${formatTime(charts.trend_chart.pb_line, 'N/A')}`, position: 'insideTopLeft', fill: '#d97706' }} stroke="#f59e0b" strokeDasharray="3 3" />}
                    </LineChart>
                  </ResponsiveContainer>
                </div>
                <div className="mt-6">
                  <h4 className="font-semibold text-slate-800 text-sm">近三場 vs PB 差距</h4>
                  <div className="mt-2 h-28">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={recentRacesVsPb} margin={{ top: 15, right: 20, left: -20, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                        <XAxis dataKey="name" fontSize={12} tickLine={false} axisLine={false} />
                        <YAxis fontSize={12} tickLine={false} axisLine={false} tickFormatter={(d) => typeof d === 'number' ? `${d.toFixed(1)}s`: ''} />
                        <RechartsTooltip formatter={(value: unknown) => typeof value === 'number' ? [`${value.toFixed(2)}s`, '差距'] : [String(value), '差距']} />
                        <ReferenceLine y={0} stroke="#475569" />
                        <Bar dataKey="diff">
                          {recentRacesVsPb?.map((entry, index) => (
                            <Bar key={`cell-${index}`} fill={entry.diff > 0 ? '#ef4444' : '#22c55e'} />
                          ))}
                        </Bar>
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </div>
              </div>

              {/* Right Panel: Insights */}
              <div className="lg:col-span-2 bg-slate-50/70 rounded-xl p-5">
                <h3 className="font-semibold text-slate-800 mb-4">💡 最近表現分析</h3>
                <div className="space-y-4 text-sm">
                  <div className="flex justify-between items-center">
                    <span className="text-slate-600">最近成績</span>
                    <span className="font-semibold">{formatTime(lastRace?.time, 'N/A')} <span className={clsx(lastRaceDiff > 0 ? 'text-red-500' : 'text-green-500', 'font-medium')}>({lastRaceDiff >= 0 ? '+' : ''}{lastRaceDiff.toFixed(2)}s vs PB)</span></span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-slate-600">近三場平均</span>
                    <span className="font-semibold">{recentThreeAvg > 0 ? `${recentThreeAvg.toFixed(2)}s` : 'N/A'} <span className={clsx(recentThreeAvgDiff > 0 ? 'text-red-500' : 'text-green-500', 'font-medium')}>({recentThreeAvgDiff >= 0 ? '+' : ''}{recentThreeAvgDiff.toFixed(2)}s vs PB)</span></span>
                  </div>
                  <hr className="border-slate-200" />
                  <div className="flex justify-between items-center">
                    <span className="text-slate-600">穩定度分數</span>
                    <span className="font-semibold text-green-600">{analysis?.stability?.value?.toFixed(0) ?? 'N/A'}<span className="text-slate-500 font-normal">/100</span></span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-slate-600">巔峰狀態評估</span>
                    <span className="font-semibold inline-flex items-center gap-1.5 bg-green-100 text-green-700 px-2 py-0.5 rounded-full">
                      <TrendingUp className="w-4 h-4" /> 接近
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
      )}

       {/* Condensed Content - shown when not expanded */}
       <div className={clsx("mt-4 space-y-3", isExpanded && "hidden")}>
            <div className="flex items-center gap-3">
              <p className="text-2xl font-bold text-amber-500 flex items-center gap-1.5">
                <Star className="w-6 h-6 fill-amber-400 text-amber-500" />
                {formatTime(personal_best?.time, 'N/A')}
              </p>
              <div className="text-xs text-slate-500">
                於 {personal_best?.date}
                <span className={clsx("ml-2 inline-flex items-center gap-1.5 font-semibold", pbFreshness.text)}>
                  <span className={clsx("w-2 h-2 rounded-full", pbFreshness.bg)}></span>
                  {pbFreshness.label}
                   <Tooltip text="此燈號反映您突破個人極限的頻率。紅燈可能表示您遇到了瓶頸期，是時候檢視訓練計畫了。">
                      <Info className="w-3.5 h-3.5 text-slate-400 cursor-pointer" />
                   </Tooltip>
                </span>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
              <div className="flex items-center gap-1.5 font-medium text-slate-600">
                穩定度:
                <span className={clsx("font-semibold flex items-center gap-1", stability.text)}>
                  {stability.label}
                  <Tooltip text="用於評估您的表現是否可靠。高穩定度代表實力已穩固；低穩定度則表示PB可能帶有偶然性。">
                    <Info className="w-3.5 h-3.5 text-slate-400 cursor-pointer" />
                   </Tooltip>
                </span>
              </div>
              <div className="flex items-center gap-1.5 font-medium text-slate-600">
                近三場趨勢:
                <span className="font-semibold flex items-center gap-1">
                  {trend.icon} {trend.label} ({analysis?.trend?.value?.toFixed(2)}s vs PB)
                   <Tooltip text="快速判斷您最近的短期狀態。箭頭顯示您是處於上升、下降還是平穩期。">
                    <Info className="w-3.5 h-3.5 text-slate-400 cursor-pointer" />
                   </Tooltip>
                </span>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
              <div className="flex items-center gap-2 font-medium text-slate-600">
                近一年表現:
                {sparklineData && sparklineData.length > 0 ? (
                  <ResponsiveContainer width={120} height={30}>
                    <LineChart data={sparklineData}>
                      <YAxis hide domain={['dataMin', 'dataMax']} tickFormatter={(time) => formatTime(Math.abs(time as number), 'N/A')} />
                      <Line type="monotone" dataKey="time" stroke="#3b82f6" strokeWidth={2} dot={false} />
                    </LineChart>
                  </ResponsiveContainer>
                ) : (
                  <div className="w-[120px] h-[30px] flex items-center justify-center text-slate-400">No data</div>
                )}
                 <Tooltip text="提供一個長期的宏觀視角，回顧過去一整年的表現起伏與賽季週期。">
                    <Info className="w-3.h-5 h-3.5 text-slate-400 cursor-pointer" />
                 </Tooltip>
              </div>
            </div>
          </div>
    </div>
  );
};

export default EventPerformanceCard;