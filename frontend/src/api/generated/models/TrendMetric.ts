/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type TrendMetric = {
    /**
     * The time difference of recent trend vs PB.
     */
    value?: number;
    unit?: string;
    /**
     * A qualitative label for the metric.
     */
    label?: TrendMetric.label;
};
export namespace TrendMetric {
    /**
     * A qualitative label for the metric.
     */
    export enum label {
        IMPROVING = 'improving',
        STABLE = 'stable',
        DECLINING = 'declining',
    }
}

