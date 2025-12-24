/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { StabilityMetric } from './StabilityMetric';
import type { TrendMetric } from './TrendMetric';
export type Analysis = {
    stability?: StabilityMetric;
    trend?: TrendMetric;
    pb_freshness?: {
        days_since_pb?: number;
        /**
         * A qualitative label for how recent the PB is.
         */
        label?: Analysis.label;
    };
};
export namespace Analysis {
    /**
     * A qualitative label for how recent the PB is.
     */
    export enum label {
        HOT_STREAK = 'hot_streak',
        STABLE = 'stable',
        NOT_UPDATED_RECENTLY = 'not_updated_recently',
    }
}

