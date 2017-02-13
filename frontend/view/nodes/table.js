(function(Vue){
  var nodesTableTemplate = `
    <div class="table-responsive">
      <table class="table table-striped jambo_table">
        <thead>
          <tr>
            <th>最新插件版本</th>
            <th>总机器数</th>
            <th>最后心跳在1h前的机器数</th>
          </tr>
        </thead>
        <tbody>
          <tr class="events-row">
            <td><img width=32 height=32 :src="'https://www.gravatar.com/avatar/' + pluginVersion + '?s=64&d=identicon&r=PG'" :alt="pluginVersion"></td>
            <td class="status">{{ numNodes }}</td>
            <td class="status">{{ numInactiveNodes }}</td>
          </tr>
        </tbody>
      </table>
      <table class="table table-striped jambo_table">
        <thead>
          <tr>
            <th>机器名</th>
            <th>IP</th>
            <th>Agent版本</th>
            <th>插件版本</th>
            <th>上次心跳</th>
            <th>插件项</th>
          </tr>
        </thead>

        <tbody>
          <tr v-if="nodes == null">
            <td colspan="6">正在获取数据……</td>
          </tr>

          <tr v-if="_.isArray(nodes) && _.isEmpty(nodes)">
            <td colspan="6">并没有机器，正确的安装 agent 了么？</td>
          </tr>

          <tr v-if="nodes"
              v-for="(n, i) in nodes"
              class="pointer events-row"
              :class="i & 1 ? 'odd' : 'even'">
            <td>{{ n.hostname }}</td>
            <td>{{ n.ip }}</td>
            <td><img width=32 height=32 :src="'https://www.gravatar.com/avatar/' + n['agent-version'].replace(/\\./g, '0') + '?s=64&d=identicon&r=PG'" :alt="n['agent-version']">{{ n['agent-version'] }}</td>
            <td><img width=32 height=32 :src="'https://www.gravatar.com/avatar/' + n['plugin-version'] + '?s=64&d=identicon&r=PG'" :alt="n['plugin-version']"></td>
            <!--<td></td>-->
            <td>{{ timeFromNow(n.lastseen) }}</td>
            <td>
              <div style="margin: 5px 0 0 5px; display: inline-block;" v-for="v in n.pluginDirs">
                <span class="label label-info">{{ v }}</span>
              </div>
              <div style="margin: 5px 0 0 5px; display: inline-block;" v-for="v in n.pluginMetrics">
                <span class="label label-warning" style="cursor: pointer" @click="showMetric(v)">{{ v._metric }}</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  `;

  var NodesTable = Vue.extend({
    template: nodesTableTemplate,
    props: ['api-endpoint', 'api-auth'],
    mounted() {
      this.refresh();
    },
    data() {
      return {
        nodes: null,
        pluginVersion: '...',
        numNodes: '...',
        numInactiveNodes: '...',
      }
    },
    methods: {
      _getFetchHeaders() {
        var headers = new Headers();
        if(this.apiAuth) {
          headers.append("Authorization", "Basic " + btoa(this.apiAuth));
        }
        return headers;
      },
      refresh() {
        var opts = {
          method: "GET",
          headers: this._getFetchHeaders(),
          credentials: 'include',
        };

        fetch(this.apiEndpoint, opts).then(resp => resp.json()).then((state) => {
          this.pluginVersion = state['plugin-version'];
          this.numNodes = Object.keys(state['agents']).length;
          this.numInactiveNodes = _.sum(_.map(state.agents, (a) => (new Date() / 1000 - a.lastseen > 3600) ? 1 : 0));
          this.nodes = _.map(state.agents, (v, k) => {
            v.pluginDirs = _.sortBy(state['plugin-dirs'][k]);
            v.pluginMetrics = _.sortBy(state['plugin-metrics'][k], v => v._metric);
            return v;
          });
        });
      },
      timeFromNow(ts) {
        return moment(new Date(ts * 1000)).locale('zh-cn').fromNow();
      },
      showMetric(m) {
        alert(JSON.stringify(m, null, 2));
      },
    }
  });

  Vue.component('nodes-table', NodesTable);
})(Vue);
