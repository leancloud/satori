.. _nvgpu:

NVidia GPU 监控
===============

这个插件提供了收集 NVidia GPU 指标的功能。

插件文件地址
    nvgpu

插件类型
    接受参数，持续执行


插件参数
--------

+----------+--------------------+
| 参数     | 功能               |
+==========+====================+
| duration | 采集周期，单位是秒 |
+----------+--------------------+


上报的监控值
------------

nvgpu.pwr
   :意义: 指定 GPU 当前功率
   :取值: 0-无上限，整数，单位是 W（瓦）
   :Tags: {"idx": "``GPU 数字编号``", "id": "``GPU ID，由型号和 GUID 的一部分组成``"},

nvgpu.gtemp
   :意义: 指定 GPU 当前内核温度
   :取值: 整数，单位是摄氏度
   :Tags: {"idx": "``GPU 数字编号``", "id": "``GPU ID，由型号和 GUID 的一部分组成``"},

nvgpu.fan
   :意义: 指定 GPU 当前风扇转速
   :取值: 0-100，百分比
   :Tags: {"idx": "``GPU 数字编号``", "id": "``GPU ID，由型号和 GUID 的一部分组成``"},

nvgpu.mem.used
   :意义: 指定 GPU 已使用显存
   :取值: 0-无上限，单位是字节
   :Tags: {"idx": "``GPU 数字编号``", "id": "``GPU ID，由型号和 GUID 的一部分组成``"},

nvgpu.mem.free
   :意义: 指定 GPU 空闲显存
   :取值: 0-无上限，单位是字节
   :Tags: {"idx": "``GPU 数字编号``", "id": "``GPU ID，由型号和 GUID 的一部分组成``"},

nvgpu.mem.total
   :意义: 指定 GPU 总显存
   :取值: 0-无上限，单位是字节
   :Tags: {"idx": "``GPU 数字编号``", "id": "``GPU ID，由型号和 GUID 的一部分组成``"},

nvgpu.util.gpu
   :意义: 指定 GPU 当前核心利用率
   :取值: 0-100，百分比
   :Tags: {"idx": "``GPU 数字编号``", "id": "``GPU ID，由型号和 GUID 的一部分组成``"},

nvgpu.util.mem
   :意义: 指定 GPU 当前显存带宽利用率
   :取值: 0-100，百分比
   :Tags: {"idx": "``GPU 数字编号``", "id": "``GPU ID，由型号和 GUID 的一部分组成``"},


监控规则样例
------------

.. code-block:: clojure

   (def nvgpu-rules
     (sdo
       (where (host "gpuhost1" "gpuhost2")
         (plugin "nvgpu" 0 {:duration 5}))

       (->waterfall
         (where (service "nvgpu.gtemp"))
         (by :host)
         (copy :id :aggregate-desc-key)
         (group-window :id)
         (aggregate max)
         (judge-gapped (> 90) (< 86))
         (alarm-every 2 :min)
         (! {:note "GPU 过热"
             :level 1
             :expected 85
             :outstanding-tags [:host]
             :groups [:operation :devs]}))

       #_(place holder)))
