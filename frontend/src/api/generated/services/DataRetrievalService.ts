/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AthleteRaceResult } from '../models/AthleteRaceResult';
import type { Competition } from '../models/Competition';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DataRetrievalService {
    /**
     * Get a list of all athletes
     * Retrieves a list of all unique athlete names present in the database.
     * @returns string A successful response returning a list of athlete names.
     * @throws ApiError
     */
    public static getAthletes(): CancelablePromise<Array<string>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/athletes',
        });
    }
    /**
     * Get a list of available competition years
     * Retrieves a list of all unique competition years present in the database.
     * @returns string A successful response returning a list of years.
     * @throws ApiError
     */
    public static getYears(): CancelablePromise<Array<string>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/years',
        });
    }
    /**
     * Get competitions for a specific year
     * Retrieves a list of competitions held in a given year.
     * @param year The year for which to retrieve competitions.
     * @param athlete Filter competitions by a single athlete name.
     * @returns Competition A successful response returning a list of competitions.
     * @throws ApiError
     */
    public static getCompetitions(
        year: string,
        athlete?: string,
    ): CancelablePromise<Array<Competition>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/competitions',
            query: {
                'year': year,
                'athlete': athlete,
            },
            errors: {
                400: `Invalid year parameter.`,
            },
        });
    }
    /**
     * Get athlete's races in a competition
     * Retrieves all race results for a specific athlete in a given competition and year.
     * @param athleteName The name of the athlete.
     * @param competitionName The name of the competition.
     * @param year The year of the competition.
     * @returns AthleteRaceResult A successful response returning a list of race results.
     * @throws ApiError
     */
    public static getAthletesRaces(
        athleteName: string,
        competitionName: string,
        year: string,
    ): CancelablePromise<Array<AthleteRaceResult>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/athletes/{athlete_name}/races',
            path: {
                'athlete_name': athleteName,
            },
            query: {
                'competition_name': competitionName,
                'year': year,
            },
            errors: {
                400: `Invalid parameters.`,
                404: `Athlete or competition not found.`,
            },
        });
    }
}
