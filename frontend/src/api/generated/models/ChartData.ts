/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type ChartData = {
    /**
     * A small array of recent results for a mini line chart.
     */
    sparkline?: Array<number>;
    trend_chart?: {
        dates?: Array<string>;
        times?: Array<number>;
        /**
         * The value of the personal best to draw a horizontal line on the chart.
         */
        pb_line?: number;
    };
};

