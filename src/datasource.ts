import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { Warp10Query, Warp10DataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<Warp10Query, Warp10DataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<Warp10DataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<Warp10Query> {
    return DEFAULT_QUERY
  }

  applyTemplateVariables(target: Warp10Query, scopedVars: ScopedVars): Record<string, any> {
    target.warpscript = getTemplateSrv().replace(target.warpscript, scopedVars);

    return target;
  }
}
