import {Panel} from '../panel/panel.js';
import {el} from '../panel/panel.js';
import {Form} from '../panel/form.js';
import {RestCall} from '../panel/rest.js';
import {AttacheExpander} from '../panel/expander.js';
import {PrettyDate} from '../panel/util.js';

"use strict";

// the controller for the Dashboard
export function FlowSingle() {
    var panel;
    var dataReq = function(){
        return {
            URL: '/flows/' + panel.IDs[0] + '/runs/' + panel.IDs[1],
        };
    }
    
    var events = [];

    // panel is view - or part of it
    var panel = new Panel(this, null, graphFlow, '#main', events, dataReq);

    this.Map = function(evt) {
        console.log("flow got a call to Map", evt);
        if (evt.Type == 'rest') {
          var pl = evt.Value.Response.Payload;
          pl.Parent = '/flows/' + panel.IDs[0];
          console.log(pl);

          pl.Graph.forEach((r, i) => {
              r.forEach((nr, ni) => {
                nr.StartedAgo = "";
                if (nr.Started != "0001-01-01T00:00:00Z") {
                    nr.StartedAgo = PrettyDate(nr.Started);
                }
                nr.Took = "";
                if (nr.Stopped != "0001-01-01T00:00:00Z") {
                    var started = new Date(nr.Started)
                    var stopped = new Date(nr.Stopped);
                    nr.Took = "("+toHHMMSS((stopped - started)/1000)+")";
                }
                pl.Graph[i][ni] = nr;
              });
          });

          console.log(pl);

          return pl;
        }
    }

    // TODO - dedupe with flow.js
    var sendData = function(data) { 
        var payload = {
            Ref: {
                ID:  panel.IDs[0],
                Ver: 1
            },
            Run: panel.IDs[1],
            Form: data
        }
        RestCall(panel.evtHub, "POST", "/push/data", payload);
    }

    /*
        Tag: "inbound.data", // will match the data types
		RunRef: event.RunRef{
			FlowRef: config.FlowRef{
				ID:  "build-project",
				Ver: 1,
			},
			Run: event.HostedIDRef{
				HostID: "h2",
				ID:     1,
			},
		},
		SourceNode: config.NodeRef{
			ID: "sign-off",
		},
		Opts: nt.Opts{
			"tests_passed": "true",
			"to_hash":      "blhahaha",
		},
        Good: true,
        */

    // AfterRender is called when the dash hs rendered containers.
    // we go and add the child summary panels
    this.AfterRender = function(data) {

        if (data == undefined) {
            return
        }
        console.log(data);
        data.Data.Graph.forEach((r, i) => {
            r.forEach((nr, ni) => {
                if (nr.Type == "data") {
                    if (nr.Enabled) {
                        console.log('draw editable form');

                        var form = {
                            ID: nr.ID,
                            fields: nr.Fields,
                        };
                        var formP = new Form('#expander-'+nr.ID, form, sendData);
                        formP.Activate();

                    } else {
                        console.log('draw uneditable values');
                    }
                }
            });
        });
        // var trigs = data.Data.Config.Triggers;
        // for (var t in trigs) {
        //     var trig = trigs[t];
        //     var form = trig.Opts.form;
        //     if (form == undefined) {
        //         continue;
        //     }
        //     // Give the form the trigger id so it can be uniquely directly referenced.
        //     form.ID = trig.ID;
        //     console.log(form);
        //     var formP = new Form('#expander-'+trig.ID, form, sendData);
        //     formP.Activate();
        // }

        AttacheExpander(el('triggers'));
        AttacheExpander(el('tasks'));
    }

    return panel;
}

function toHHMMSS(sec_num) {
    sec_num = Math.floor(sec_num)
    var hours   = Math.floor(sec_num / 3600);
    var minutes = Math.floor((sec_num - (hours * 3600)) / 60);
    var seconds = sec_num - (hours * 3600) - (minutes * 60);

    
    if (minutes < 10) {minutes = "0"+minutes;}
    if (seconds < 10) {seconds = "0"+seconds;}
    if (hours > 0) {
        return hours+':'+minutes+':'+seconds;
    }
    return minutes+':'+seconds;
}

var graphFlow = `
    <div id='flow' class='flow-single'>
        <div class="crumb">
          <a href='{{=it.Data.Parent}}'>← back to {{=it.Data.FlowName}}</a>
        </div>
        <summary>
            <h3>{{=it.Data.Name}}</a></h3>
        </summary>
        <triggers>
          {{~it.Data.Triggers :trigger:index}}

          <box id='trig-{{=trigger.ID}}' class='trigger{{? !trigger.Enabled}} disabled{{?}}'>
              {{? trigger.Type=='data'}}
              <div for="{{=trigger.ID}}" class="data-title expander-ctrl">
                  <h4>{{=trigger.Name}}</h4><i class='icon-angle-circled-right'></i>
              </div>
              {{??}}
              <div class="data-title">
                  <h4>{{=trigger.Name}}</h4>
              </div>
              {{?}}
              {{? trigger.Type=='data'}}
              <detail id='expander-{{=trigger.ID}}' class='expander'>
                  {{~trigger.Fields :field:index}}
                    <div id="field-{{=field.id}}", class='kvrow'>
                      <div class='prompt'>{{=field.prompt}}:</div> 
                      <div class='value'>{{=field.value}}</div>
                    </div>
                  {{~}}
              </detail>
              {{?}}
          </box>

          {{~}}
        </triggers>
        <divider></divider>
        <tasks>
        {{~it.Data.Graph :level:index}}
          <div id='level-{{=index}}' class='level'>
          {{~level :node:indx}}
            <box id='node-{{=node.ID}}' class='task {{=node.Result}} {{=node.Status}}'>
              {{? node.Type=="data"}}
              <div for="{{=node.ID}}" class="data-title expander-ctrl">
                  <h4>{{=node.Name}}</h4><i class='icon-angle-circled-right'></i>
              </div>
              <detail id='expander-{{=node.ID}}' class='expander'>
                 <section class='trig-form'></section>
              </detail>
              {{??}}
              <h4>{{=node.Name}}</h4>
              <detail>
                <p class='ago'>{{=node.StartedAgo}}</p><p class='took'>{{=node.Took}}</p>
              <detail>
              {{?}}
            </box>
          {{~}}
          </div>
        {{~}}
        </tasks>
    </div>
`