import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface Warp10Query extends DataQuery {
  warpscript?: string;
  alias?: string;
  showLabels: boolean;
}

export const DEFAULT_QUERY: Partial<Warp10Query> = {};

/**
 * These are options configured for each DataSource instance
 */
export interface Warp10DataSourceOptions extends DataSourceJsonData {
  baseUrl?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface Warp10SecureJsonData {
  token?: string;
}
