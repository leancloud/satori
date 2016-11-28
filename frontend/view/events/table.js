(function(Vue){
  var evTableTemplate = `
    <div class="table-responsive">
      <table class="table table-striped jambo_table bulk_action">
        <thead>
          <tr>
            <!--<th></th>-->
            <th>状态</th>
            <th>级别</th>
            <th>节点</th>
            <th>事件</th>
            <th>通知组</th>
            <th>监控项</th>
            <th>监控值</th>
            <th>触发时间</th>
            <th>标签</th>
            <th class="no-link last"><span class="nobr">动作</span></th>
          </tr>
        </thead>

        <tbody>
          <tr v-if="alarms == null">
            <td colspan="10">正在获取数据……</td>
          </tr>

          <tr v-if="_.isArray(alarms) && _.isEmpty(alarms)">
            <td colspan="10">并没有什么大新闻😆</td>
          </tr>

          <tr v-if="alarms"
              v-for="(a, i) in alarms"
              :key="a.id"
              class="pointer events-row"
              :class="[i & 1 ? 'odd' : 'even', _.includes(checked, a.id) ? 'selected' : '']">
            <!--
            <td class="a-center">
              <input type="checkbox" :value="a.id" class="icheck" v-model="checked">
            </td>
            -->
            <td class="status" :title="stateDescription(a.status)">{{ stateEmoji(a.status) }}</td>
            <td class="status">{{ ['0⃣','1⃣','2⃣','3⃣','4⃣','5⃣','6⃣','7⃣','8⃣','9⃣'][parseInt(a.level)] }}</td>

            <td>{{ a.endpoint }}</td>
            <td>{{ a.note }}</td>
            <td><span class="label label-primary" style="margin: 0 3px 0 3px;" v-for="g in a.groups">{{ g }}</span></td>
            <td>{{ a.metric }}</td>
            <td>{{ _.round(a.actual, 3) }}</td>
            <td>{{ timeFromNow(a.time) }}</td>
            <td>
              <div style="margin: 5px 0 0 5px; display: inline-block;" v-for="(v, k) in a.tags">
                <span class="label label-warning no-right-radius">{{ k }}</span><span class="label label-info no-left-radius">{{ v }}</span>
              </div>
            </td>
            <td class="last">
              <button v-show="a.status == 'PROBLEM'" @click="toggleAck(a)" class="btn btn-warning">静音</button>
              <button v-show="a.status == 'ACK'" @click="toggleAck(a)" class="btn btn-info">解除静音</button>
              <button @click="remove(a, i)" class="btn btn-danger">删除</button>
            </td>
          </tr>
          <!--
          <tr class="pointer events-row">
            <td>
              <input type="checkbox" class="icheck" v-model="checkAll" @change="doCheckAll()">
            </td>
            <td colspan="10">
              <a class="antoo" style="font-weight:500;">批量操作 ( {{ checked.length }} 条记录)</a>
              <div style="display: inline-block; margin-left: 10px;">
                <button @click="batchAck(true)" class="btn btn-warning">静音</button>
                <button @click="batchAck(false)" class="btn btn-info">解除静音</button>
                <button @click="batchRemove()" class="btn btn-danger">删除</button>
              </div>
            </td>
          </tr>
          -->
        </tbody>
      </table>
    </div>
  `;

  var EventsTable = Vue.extend({
    template: evTableTemplate,
    props: ['api-endpoint', 'api-auth'],
    mounted() {
      this.refresh();
    },
    data() {
      return {
        checkAll: false,
        alarms: null,
        checked: [],
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

        var order = {
          'PROBLEM': 'AAAA',
          'FLAPPING': 'BBBB',
          'ACK': 'CCCC',
          'TIMEWAIT': 'ZZZZ',
        };

        fetch(this.apiEndpoint, opts).then(resp => resp.json()).then((data) => {
          this.alarms = _.sortBy(data['alarms'], a => (order[a.status] + a.title));
        });
      },
      stateEmoji(s) {
        var emoji = {
          "PROBLEM": "😱",
          "ACK": "🔕",
          "FLAPPING": "🎭", // 🔃  🔄
          "TIMEWAIT": "⌛",
          "ERROR": "❌",
        }[s];
        return emoji ? emoji : s;
      },
      stateDescription(s) {
        var desc = {
          "PROBLEM": "现在存在的问题",
          "ACK": "静音的问题",
          "FLAPPING": "被静音后不停重复发生的问题", // 🔃  🔄
          "TIMEWAIT": "被静音后解决了的问题（在观察期内，超时后会自己消失）",
          "ERROR": "错误",
        }[s];
        return desc ? desc : s;
      },
      toggleAck(item) {
        var opts = {
          method: "POST",
          headers: this._getFetchHeaders(),
        };

        fetch(`${ this.apiEndpoint }/${ item.id }/toggle-ack`, opts).then(resp => resp.json()).then((data) => {
          item.status = data["new-state"];
        });
      },
      batchAck(state) {
        _.each(this.alarms, (a) => {
          if((state == true && a.status == 'PROBLEM') ||
             (state == false && a.status == 'ACK')) {
            if(_.includes(this.checked, a.id)) {
              this.toggleAck(a);
            }
          }
        });
      },
      remove(item, index) {
        var opts = {
          method: "DELETE",
          headers: this._getFetchHeaders(),
        };

        //*
        fetch(`${ this.apiEndpoint }/${ item.id }`, opts).then(resp => resp.json()).then((data) => {
          this.alarms.splice(index, 1);
        });
        // */
      },
      batchRemove() {
        // meh
      },
      doCheckAll() {
        if(this.checkAll) {
          this.checked = _.map(this.alarms, (i) => i.id);
        } else {
          this.checked = [];
        }
      },
      timeFromNow(ts) {
        return moment(new Date(ts * 1000)).locale('zh-cn').fromNow();
      },
    }
  });

  Vue.component('events-table', EventsTable);
})(Vue);
