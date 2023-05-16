import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { Warp10Query, Warp10DataSourceOptions, DEFAULT_QUERY } from './types';

import { Warp10VariableSupport } from './variables';

export class DataSource extends DataSourceWithBackend<Warp10Query, Warp10DataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<Warp10DataSourceOptions>) {
    super(instanceSettings);

    this.variables = new Warp10VariableSupport(this);
  }

  getDefaultQuery(_: CoreApp): Partial<Warp10Query> {
    return DEFAULT_QUERY;
  }

  getQueryDisplayText(query: Warp10Query) {
    return "query definiton";
  }

  applyTemplateVariables(target: Warp10Query, scopedVars: ScopedVars): Record<string, any> {
    if (target.warpscript !== undefined) {
      target.warpscript = getTemplateSrv().replace(target.warpscript, scopedVars, (data: string | string[]) => {
        if (Array.isArray(data)) {
          if (data.length === 1) {
            return `'${data[0]}'`;
          }

          const concatainsValues = data.join("|");

          return `'~(${concatainsValues})'`;
        }

        return `'${data}'`;
      });
    }

    return target;
  }
}
