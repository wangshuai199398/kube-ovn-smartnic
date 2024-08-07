#!/bin/bash

set -eo pipefail
OVN_REMOTE_PROBE_INTERVAL=${OVN_REMOTE_PROBE_INTERVAL:-10000}
OVN_REMOTE_OPENFLOW_INTERVAL=${OVN_REMOTE_OPENFLOW_INTERVAL:-180}

echo "OVN_REMOTE_PROBE_INTERVAL is set to $OVN_REMOTE_PROBE_INTERVAL"
echo "OVN_REMOTE_OPENFLOW_INTERVAL is set to $OVN_REMOTE_OPENFLOW_INTERVAL"

DPDK_TUNNEL_IFACE=${DPDK_TUNNEL_IFACE:-br-phy}
TUNNEL_TYPE=${TUNNEL_TYPE:-geneve}

OVS_DPDK_CONFIG_FILE=/opt/ovs-config/ovs-dpdk-config
if ! test -f "$OVS_DPDK_CONFIG_FILE"; then
    echo "missing ovs dpdk config"
    exit 1
fi
source $OVS_DPDK_CONFIG_FILE

# link sock
mkdir -p /usr/local/var/run

if [ -L /usr/local/var/run/openvswitch ]
then
     echo "sock exist"
else
     echo "link sock"
     ln -s /var/run/openvswitch /usr/local/var/run/openvswitch
fi

export PATH=$PATH:/usr/share/ovn/scripts

function quit {
	ovn-ctl stop_controller
	exit 0
}
trap quit EXIT

ovs-vsctl --may-exist add-br br-int \
  -- set Bridge br-int datapath_type=netdev \
  -- br-set-external-id br-int bridge-id br-int \
  -- set bridge br-int fail-mode=secure



# Start ovn-controller
ovn-ctl restart_controller

# Set remote ovn-sb for ovn-controller to connect to
ovs-vsctl set open . external-ids:ovn-remote=tcp:"${OVN_SB_SERVICE_HOST}":"${OVN_SB_SERVICE_PORT}"
ovs-vsctl set open . external-ids:ovn-remote-probe-interval="${OVN_REMOTE_PROBE_INTERVAL}"
ovs-vsctl set open . external-ids:ovn-openflow-probe-interval="${OVN_REMOTE_OPENFLOW_INTERVAL}"
ovs-vsctl set open . external-ids:ovn-encap-type="${TUNNEL_TYPE}"

tail --follow=name --retry /var/log/openvswitch/ovs-vswitchd.log