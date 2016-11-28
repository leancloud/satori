(function(Vue){
  var sidebarTemplate = `
    <div class="col-md-3 left_col">
      <div class="left_col scroll-view">
        <div class="navbar nav_title" style="border: 0;">
          <a href="index.html" class="site_title"><i class="fa fa-eye"></i> <span>SATORI</span></a>
        </div>
        <div class="clearfix"></div>
        <br />
        <!-- sidebar menu -->
        <div id="sidebar-menu" class="main_menu_side hidden-print main_menu">
          <div class="menu_section">
            <!--<h3>General</h3>-->
            <ul class="nav side-menu">
              <li v-for="item in items" :class="current === item.id ? 'active-sm' : ''">
                <a :href="item.url"><i class = "fa" :class="item.icon"></i> {{ item.text }}</a>
              </li>
            </ul>
          </div>
        </div>
        <!-- /sidebar menu -->
      </div>
    </div>
  `;

  var Sidebar = Vue.extend({
    template: sidebarTemplate,
    props: ['current'],
    data: function() {
      return {
        items: [
          { id: 'index',    url: 'index.html',  icon: 'fa-home',       text: '首页'},
          { id: 'events',   url: 'events.html', icon: 'fa-bell',       text: '报警'},
          { id: 'nodes',    url: 'nodes.html',  icon: 'fa-server',     text: '机器信息'},
          { id: 'grafana',  url: 'grafana/',    icon: 'fa-area-chart', text: '图表'},
          // { id: 'influxdb', url: 'influxdb/',    icon: 'fa-database',   text: 'InfluxDB'},
        ],
      };
    },
    ready: function() {
    },
  });

  Vue.component('sidebar', Sidebar);
})(Vue);
