/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { EventPerformance } from '../models/EventPerformance';
import type { ResultComparison } from '../models/ResultComparison';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class PerformanceService {
    /**
     * Get performance overview for an athlete
     * Retrieves a detailed performance analysis for a specific athlete, broken down by event.
     * It combines raw statistics with calculated metrics like stability, trend, and PB freshness.
     *
     * @param athleteName The name of the athlete to retrieve data for.
     * @returns EventPerformance A successful response returning the performance overview.
     * @throws ApiError
     */
    public static getAthletesPerformanceOverview(
        athleteName: string,
    ): CancelablePromise<Array<EventPerformance>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/athletes/{athlete_name}/performance-overview',
            path: {
                'athlete_name': athleteName,
            },
            errors: {
                404: `Athlete not found.`,
            },
        });
    }
    /**
     * Get comparison for a single race
     * Retrieves a detailed comparison for a specific race, pitting a target athlete's result against other competitors in the same race and against the national and games records.
     *
     * @param raceId The ID of the race to compare.
     * @param athleteName The name of the athlete to filter the comparison by.
     * @returns ResultComparison A successful response returning the result comparison.
     * @throws ApiError
     */
    public static getRaceComparison(
        raceId: string,
        athleteName: string,
    ): CancelablePromise<ResultComparison> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/race/{race_id}/comparison',
            path: {
                'race_id': raceId,
            },
            query: {
                'athlete_name': athleteName,
            },
            errors: {
                404: `Race not found.`,
            },
        });
    }
}
