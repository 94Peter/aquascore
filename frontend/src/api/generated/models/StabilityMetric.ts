/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type StabilityMetric = {
    /**
     * The coefficient of variation for stability.
     */
    value?: number;
    unit?: string;
    /**
     * A qualitative label for the metric.
     */
    label?: StabilityMetric.label;
};
export namespace StabilityMetric {
    /**
     * A qualitative label for the metric.
     */
    export enum label {
        HIGH = 'high',
        MEDIUM = 'medium',
        LOW = 'low',
    }
}

