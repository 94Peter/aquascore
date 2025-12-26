import pandas as pd
from datetime import datetime, timedelta
from google.protobuf.timestamp_pb2 import Timestamp
from analysis.v1 import analysis_pb2

# Helper functions for labels and time conversions (based on wireframe/openapi)
def get_pb_freshness_label(days_since_pb: int) -> str:
    if days_since_pb <= 90:  # 3 months
        return "hot_streak"
    elif days_since_pb <= 180: # 6 months
        return "stable"
    else:
        return "not_updated_recently"

def get_stability_label(cv_value: float) -> str:
    # Assuming lower CV means higher stability. Thresholds are examples.
    if cv_value <= 5.0:
        return "high"
    elif cv_value <= 10.0:
        return "medium"
    else:
        return "low"

def get_trend_label(trend_value: float) -> str:
    # trend_value is (average of last 3 races) - PB. Negative means improving.
    if trend_value < -0.1: # Significantly improving
        return "improving"
    elif trend_value > 0.1: # Significantly declining
        return "declining"
    else:
        return "stable"

def format_date_to_string(dt: datetime) -> str:
    return dt.strftime("%Y-%m-%d")

def analyze_performance_overview(results: list) -> list:
    """
    Analyzes performance results for an athlete, calculating metrics for each event.
    """
    if not results:
        return []

    # 1. Convert protobuf messages to a list of dictionaries for DataFrame creation
    # Also extract competition name if available from PerformanceResult
    # If competition_name is needed for RecentRace, it must be added to PerformanceResult.
    # For now, I'll assume we can use a placeholder for RecentRace.competition_name
    data = []
    for res in results:
        # Assuming PerformanceResult might have a competition_name field if needed for recent_races
        # If not, it needs to be passed in the request or derived.
        # For the purpose of this task, I'll assume we can use a placeholder for RecentRace.competition_name
        data.append({
            "date": res.event_date.ToDatetime(), # Keep as datetime for easier calculations
            "record": res.result_time,
            "event_type": res.event_type,
            # Placeholder for competition_name, ideally this comes from PerformanceResult
            "competition_name": res.competition_name,
        })


    if not data:
        return []

    # 2. Create and process the DataFrame
    df = pd.DataFrame(data)
    df['date'] = pd.to_datetime(df['date']) # Ensure datetime type
    df = df.sort_values(by="date")

    event_analyses = [] # Rename to event_analyses for clarity with protobuf response

    # 3. Group by event and calculate metrics for each
    for event_name, group_df in df.groupby('event_type'):
        if group_df.empty:
            continue

        # Calculate Personal Best (PB)
        pb_row = group_df.loc[group_df['record'].idxmin()]
        pb_time = pb_row['record']
        pb_date_dt = pb_row['date'] # datetime object

        # Construct PersonalBest message
        personal_best_msg = analysis_pb2.PersonalBest(
            time=pb_time,
            unit="s", # Assuming seconds for swimming events
            date=format_date_to_string(pb_date_dt)
        )

        # Calculate PB Freshness
        days_since_pb = (datetime.now() - pb_date_dt).days
        pb_freshness_msg = analysis_pb2.PBFreshness(
            days_since_pb=days_since_pb,
            label=get_pb_freshness_label(days_since_pb)
        )

        # Calculate stability (Coefficient of Variation) for the last 5 races
        last_5_races = group_df.tail(5)
        stability_value = 0.0
        if len(last_5_races) > 1:
            mean_val = last_5_races['record'].mean()
            if mean_val > 0:
                stability_value = round((last_5_races['record'].std(ddof=0) / mean_val) * 100, 2)
        stability_msg = analysis_pb2.StabilityMetric(
            value=stability_value,
            unit="%",
            label=get_stability_label(stability_value)
        )

        # Calculate recent trend (last race vs. oldest race in the last 3)
        last_3_races = group_df.tail(3)
        trend_value = 0.0
        if len(last_3_races) >= 2: # Need at least 2 races to determine a trend in the sequence
            # Compare the last race to the first in the window
            trend_value = round(last_3_races['record'].iloc[-1] - last_3_races['record'].iloc[0], 2)
        elif len(last_3_races) == 1:
            # If only one race, compare it to PB (this might need further refinement based on requirements)
            trend_value = round(last_3_races['record'].iloc[-1] - pb_time, 2)
        # else: trend_value remains 0.0 (stable)
        trend_msg = analysis_pb2.TrendMetric(
            value=trend_value,
            unit="s",
            label=get_trend_label(trend_value)
        )

        # Construct AnalysisMetrics message
        analysis_metrics_msg = analysis_pb2.AnalysisMetrics(
            stability=stability_msg,
            trend=trend_msg,
            pb_freshness=pb_freshness_msg
        )

        # Construct RecentRaces
        # Take the last few races, for example, up to 5 recent races
        recent_races_list = []
        for _, row in group_df.tail(5).iterrows(): # Using tail(5) for recent races
            recent_races_list.append(analysis_pb2.RecentRace(
                date=format_date_to_string(row['date']),
                time=row['record'],
                competition_name=row['competition_name'] # Use actual competition name if available
            ))

        # Construct Charts
        sparkline_data = group_df['record'].tail(12).tolist() # last 12 for sparkline
        
        # For Trend Chart, use a larger history, e.g., last 10 races
        trend_chart_history = group_df.tail(10)
        trend_chart_msg = analysis_pb2.TrendChart(
            dates=[format_date_to_string(d) for d in trend_chart_history['date'].tolist()],
            times=trend_chart_history['record'].tolist(),
            pb_line=pb_time
        )
        chart_data_msg = analysis_pb2.ChartData(
            sparkline=sparkline_data,
            trend_chart=trend_chart_msg
        )

        # Assemble the full EventPerformanceAnalysis result for the event
        event_analysis_msg = analysis_pb2.EventPerformanceAnalysis(
            event_name=event_name,
            personal_best=personal_best_msg,
            analysis=analysis_metrics_msg,
            recent_races=recent_races_list,
            charts=chart_data_msg
        )
        event_analyses.append(event_analysis_msg)

    return event_analyses