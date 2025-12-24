/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Analysis } from './Analysis';
import type { ChartData } from './ChartData';
import type { PersonalBest } from './PersonalBest';
import type { RecentRace } from './RecentRace';
export type EventPerformance = {
    /**
     * The name of the swimming event (e.g., "50m Freestyle").
     */
    event_name?: string;
    personal_best?: PersonalBest;
    analysis?: Analysis;
    recent_races?: Array<RecentRace>;
    charts?: ChartData;
};

