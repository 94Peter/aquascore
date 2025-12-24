from analysis.v1 import analysis_pb2

def analyze_result_comparison(target_result: analysis_pb2.RaceResult, competition_results: list, records: analysis_pb2.RecordMarks) -> list:
    """
    Analyzes a target result against other results and records.
    """
    comparisons = []

    # Combine target result with competition results if it's not already there
    all_results_dict = {res.athlete_name: res for res in competition_results}
    all_results_dict[target_result.athlete_name] = target_result
    
    all_results = list(all_results_dict.values())

    for result in all_results:
        # Create the base comparison object
        comp = analysis_pb2.SingleResultComparison(
            athlete_name=result.athlete_name,
            record_time=result.record_time,
            rank=result.rank,
        )

        # Calculate difference from national record if available
        if records.HasField("national_record"):
            comp.diff_from_national_record = round(result.record_time - records.national_record, 2)

        # Calculate difference from games record if available
        if records.HasField("games_record"):
            comp.diff_from_games_record = round(result.record_time - records.games_record, 2)

        # Calculate difference from target athlete
        if result.athlete_name != target_result.athlete_name:
            comp.diff_from_target = round(result.record_time - target_result.record_time, 2)

        comparisons.append(comp)
    
    # Sort results by rank to ensure a consistent order
    comparisons.sort(key=lambda x: x.rank)

    return comparisons
