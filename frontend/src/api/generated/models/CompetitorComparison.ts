/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type CompetitorComparison = {
    rank?: number;
    athlete_name?: string;
    record_time?: number;
    /**
     * The time difference from the target athlete.
     */
    diff_from_target?: string;
    /**
     * A qualitative label for the time difference.
     */
    diff_label?: CompetitorComparison.diff_label;
};
export namespace CompetitorComparison {
    /**
     * A qualitative label for the time difference.
     */
    export enum diff_label {
        FAR_AHEAD = 'far_ahead',
        SLIGHTLY_AHEAD = 'slightly_ahead',
        YOUR_RESULT = 'your_result',
        SLIGHTLY_BEHIND = 'slightly_behind',
        FAR_BEHIND = 'far_behind',
    }
}

