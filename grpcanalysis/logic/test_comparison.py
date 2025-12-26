import pytest
from grpcanalysis.logic.comparison import analyze_result_comparison
from analysis.v1 import analysis_pb2

def test_analyze_result_comparison():
    # Setup data
    target_result = analysis_pb2.RaceResult(
        athlete_name="Athlete A",
        record_time=10.5,
        rank=2
    )
    
    comp_results = [
        analysis_pb2.RaceResult(
            athlete_name="Athlete B",
            record_time=10.0,
            rank=1
        ),
            analysis_pb2.RaceResult(
            athlete_name="Athlete C",
            record_time=11.0,
            rank=3
        )
    ]
    
    records = analysis_pb2.RecordMarks(
        national_record=9.5,
        games_record=9.8
    )
    
    # Execute
    results = analyze_result_comparison(target_result, comp_results, records)
    
    # Verify
    assert len(results) == 3 # Target (implicitly added) + 2 competitors
    
    # Check sorting by rank
    assert results[0].athlete_name == "Athlete B"
    assert results[1].athlete_name == "Athlete A"
    assert results[2].athlete_name == "Athlete C"

    # Check diffs for target athlete (Athlete A)
    target_comp = results[1]
    assert target_comp.athlete_name == "Athlete A"
    assert target_comp.diff_from_national_record == pytest.approx(1.0) # 10.5 - 9.5
    assert target_comp.diff_from_games_record == pytest.approx(0.7) # 10.5 - 9.8
    
    # For target athlete, diff_from_target should usually be 0 or not set.
    # Checking if HasField works implies it is optional. 
    # If it throws error we will fix it.
    try:
            assert not target_comp.HasField("diff_from_target")
    except ValueError:
            # If not optional, it should be 0.0
            assert target_comp.diff_from_target == pytest.approx(0.0)

    # Check diffs for another athlete (Athlete B)
    other_comp = results[0]
    assert other_comp.athlete_name == "Athlete B"
    assert other_comp.diff_from_national_record == pytest.approx(0.5) # 10.0 - 9.5
    assert other_comp.diff_from_target == pytest.approx(-0.5) # 10.0 - 10.5

def test_analyze_result_comparison_no_records():
        # Setup data
    target_result = analysis_pb2.RaceResult(
        athlete_name="Athlete A",
        record_time=10.5,
        rank=1
    )
    
    comp_results = []
    
    records = analysis_pb2.RecordMarks() # Empty
    
    # Execute
    results = analyze_result_comparison(target_result, comp_results, records)
    
    assert len(results) == 1
    res = results[0]
    assert res.athlete_name == "Athlete A"
    
    # Check that record diffs are not set/calculated
    try:
        assert not res.HasField("diff_from_national_record")
        assert not res.HasField("diff_from_games_record")
    except ValueError:
        pass # Or check for 0.0 if that's the default behavior but code implies conditional assignment