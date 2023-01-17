



struct ethernet_t {
	bit<48> dst_addr
	bit<48> src_addr
	bit<16> ether_type
}

struct arp_t {
	bit<16> hw_type
	bit<16> proto_type
	bit<8> hw_addr_len
	bit<8> proto_addr_len
	bit<16> opcode
	bit<48> sender_hw_addr
	bit<32> sender_proto_addr
	bit<48> target_hw_addr
	bit<32> target_proto_addr
}

struct ipv4_t {
	bit<8> version_ihl
	bit<8> dscp_ecn
	bit<16> total_len
	bit<16> identification
	bit<16> flags_frag_offset
	bit<8> ttl
	bit<8> protocol
	bit<16> header_checksum
	bit<32> src_addr
	bit<32> dst_addr
}

struct tcp_t {
	bit<16> src_port
	bit<16> dst_port
	bit<32> seq_no
	bit<32> ack_no
	bit<8> data_offset_res
	bit<8> flags
	bit<16> window
	bit<16> checksum
	bit<16> urgent_ptr
}

struct udp_t {
	bit<16> src_port
	bit<16> dst_port
	bit<16> hdr_length
	bit<16> checksum
}

struct icmp_t {
	bit<8> type
	bit<8> code
	bit<16> checksum
}

struct vxlan_t {
	bit<8> flags
	bit<24> reserved
	bit<24> vni
	bit<8> reserved2
}

struct cksum_state_t {
	bit<16> state_0
}

struct vlan_t {
	bit<16> pcp_cfi_vid
	bit<16> ether_type
}

struct decap_outer_ipv4_arg_t {
	bit<24> tunnel_id
}

struct ipv4_table_set_group_id_arg_t {
	bit<32> group_id
}

struct ipv4_table_set_member_id_arg_t {
	bit<32> member_id
}

struct l2_fwd_1_arg_t {
	bit<32> port
}

struct l2_fwd_2_arg_t {
	bit<32> port
}

struct l2_fwd_arg_t {
	bit<32> port
}

struct pop_vlan_fwd_arg_t {
	bit<32> port
}

struct push_vlan_fwd_arg_t {
	bit<32> port
	bit<16> vlan_tag
}

struct set_control_dest_1_arg_t {
	bit<32> port_id
}

struct set_control_dest_arg_t {
	bit<32> port_id
}

struct set_nexthop_arg_t {
	bit<16> router_interface_id
	bit<24> neighbor_id
	bit<32> egress_port
}

struct set_nexthop_id_arg_t {
	bit<16> nexthop_id
}

struct set_outer_mac_arg_t {
	bit<48> dst_mac_addr
}

struct set_src_mac_arg_t {
	bit<48> src_mac_addr
}

struct set_tunnel_arg_t {
	bit<24> tunnel_id
	bit<32> dst_addr
}

struct vxlan_encap_arg_t {
	bit<32> src_addr
	bit<32> dst_addr
	bit<16> dst_port
	bit<24> vni
}

header outer_ethernet instanceof ethernet_t
header outer_vlan_0 instanceof vlan_t
header outer_vlan_1 instanceof vlan_t

header outer_arp instanceof arp_t
header outer_ipv4 instanceof ipv4_t
header outer_tcp instanceof tcp_t
header outer_udp instanceof udp_t
header outer_icmp instanceof icmp_t
header vxlan instanceof vxlan_t
header ethernet instanceof ethernet_t
header vlan_0 instanceof vlan_t
header vlan_1 instanceof vlan_t

header arp instanceof arp_t
header ipv4 instanceof ipv4_t
header udp instanceof udp_t
header tcp instanceof tcp_t
header icmp instanceof icmp_t
header cksum_state instanceof cksum_state_t

struct local_metadata_t {
	bit<32> pna_main_input_metadata_direction
	bit<32> pna_main_input_metadata_input_port
	bit<8> local_metadata__control_packet1
	bit<16> local_metadata__nexthop_id6
	bit<16> local_metadata__rif_mod_map_id7
	bit<16> local_metadata__vlan_id8
	bit<8> local_metadata__is_tunnel9
	bit<16> local_metadata__hash10
	bit<32> local_metadata__ipv4_dst_match11
	bit<24> local_metadata__tunnel_id12
	bit<8> local_metadata__tunnel_tun_type14
	bit<32> pna_main_output_metadata_output_port
	;oldname:linux_networking_control_ipv4_tunnel_term_table_outer_ipv4_src_addr
	bit<32> linux_networking_control_ipv4_tunnel_term_table_outer_ipv4_0
	;oldname:linux_networking_control_ipv4_tunnel_term_table_outer_ipv4_dst_addr
	bit<32> linux_networking_control_ipv4_tunnel_term_table_outer_ipv4_1
	bit<32> linux_networking_control_as_ecmp_sel_outer_ipv4_src_addr
	bit<32> linux_networking_control_as_ecmp_sel_outer_ipv4_dst_addr
	bit<8> linux_networking_control_as_ecmp_sel_outer_ipv4_protocol
	bit<16> linux_networking_control_as_ecmp_sel_outer_udp_src_port
	bit<16> linux_networking_control_as_ecmp_sel_outer_udp_dst_port
	bit<16> MainControlT_tmp
	bit<16> MainControlT_tmp_0
	bit<48> MainControlT_tmp_1
	bit<48> MainControlT_tmp_2
	bit<16> MainControlT_vendormeta_mod_action_ref
	bit<24> MainControlT_vendormeta_mod_data_ptr
	bit<24> MainControlT_vendormeta_neighbor_mod_data_ptr
	bit<32> MainControlT_as_ecmp_group_id
	bit<32> MainControlT_as_ecmp_member_id
}
metadata instanceof local_metadata_t

regarray direction size 0x100 initval 0

action NoAction args none {
	return
}

action drop args none {
	drop
	return
}

action drop_1 args none {
	drop
	return
}

action drop_2 args none {
	drop
	return
}

action drop_3 args none {
	drop
	return
}

action drop_4 args none {
	drop
	return
}

action drop_5 args none {
	drop
	return
}

action drop_6 args none {
	drop
	return
}

action drop_7 args none {
	drop
	return
}

action drop_8 args none {
	drop
	return
}

action vxlan_encap args instanceof vxlan_encap_arg_t {
	validate h.ethernet
	mov h.ethernet.dst_addr h.outer_ethernet.dst_addr
	mov h.ethernet.src_addr h.outer_ethernet.src_addr
	mov h.ethernet.ether_type h.outer_ethernet.ether_type
	invalidate h.outer_ethernet
	jmpnv LABEL_END_7 h.outer_vlan_0
	mov h.vlan_0.pcp_cfi_vid h.outer_vlan_0.pcp_cfi_vid
	mov h.vlan_0.ether_type h.outer_vlan_0.ether_type
	invalidate h.outer_vlan_0
	LABEL_END_7 :	jmpnv LABEL_END_8 h.outer_vlan_1
	mov h.vlan_1.pcp_cfi_vid h.outer_vlan_1.pcp_cfi_vid
	mov h.vlan_1.ether_type h.outer_vlan_1.ether_type
	invalidate h.outer_vlan_1
	LABEL_END_8 :	jmpnv LABEL_END_9 h.outer_arp
	mov h.arp.hw_type h.outer_arp.hw_type
	mov h.arp.proto_type h.outer_arp.proto_type
	mov h.arp.hw_addr_len h.outer_arp.hw_addr_len
	mov h.arp.proto_addr_len h.outer_arp.proto_addr_len
	mov h.arp.opcode h.outer_arp.opcode
	mov h.arp.sender_hw_addr h.outer_arp.sender_hw_addr
	mov h.arp.sender_proto_addr h.outer_arp.sender_proto_addr
	mov h.arp.target_hw_addr h.outer_arp.target_hw_addr
	mov h.arp.target_proto_addr h.outer_arp.target_proto_addr
	invalidate h.outer_arp
	LABEL_END_9 :	jmpnv LABEL_END_10 h.outer_ipv4
	validate h.ipv4
	mov h.ipv4.version_ihl h.outer_ipv4.version_ihl
	mov h.ipv4.dscp_ecn h.outer_ipv4.dscp_ecn
	mov h.ipv4.total_len h.outer_ipv4.total_len
	mov h.ipv4.identification h.outer_ipv4.identification
	mov h.ipv4.flags_frag_offset h.outer_ipv4.flags_frag_offset
	mov h.ipv4.ttl h.outer_ipv4.ttl
	mov h.ipv4.protocol h.outer_ipv4.protocol
	mov h.ipv4.header_checksum h.outer_ipv4.header_checksum
	mov h.ipv4.src_addr h.outer_ipv4.src_addr
	mov h.ipv4.dst_addr h.outer_ipv4.dst_addr
	invalidate h.outer_ipv4
	LABEL_END_10 :	jmpnv LABEL_END_11 h.outer_icmp
	validate h.icmp
	mov h.icmp.type h.outer_icmp.type
	mov h.icmp.code h.outer_icmp.code
	mov h.icmp.checksum h.outer_icmp.checksum
	invalidate h.outer_icmp
	LABEL_END_11 :	jmpnv LABEL_END_12 h.outer_udp
	validate h.udp
	mov h.udp.src_port h.outer_udp.src_port
	mov h.udp.dst_port h.outer_udp.dst_port
	mov h.udp.hdr_length h.outer_udp.hdr_length
	mov h.udp.checksum h.outer_udp.checksum
	invalidate h.outer_udp
	LABEL_END_12 :	jmpnv LABEL_END_13 h.outer_tcp
	validate h.tcp
	mov h.tcp.src_port h.outer_tcp.src_port
	mov h.tcp.dst_port h.outer_tcp.dst_port
	mov h.tcp.seq_no h.outer_tcp.seq_no
	mov h.tcp.ack_no h.outer_tcp.ack_no
	mov h.tcp.data_offset_res h.outer_tcp.data_offset_res
	mov h.tcp.flags h.outer_tcp.flags
	mov h.tcp.window h.outer_tcp.window
	mov h.tcp.checksum h.outer_tcp.checksum
	mov h.tcp.urgent_ptr h.outer_tcp.urgent_ptr
	invalidate h.outer_tcp
	LABEL_END_13 :	validate h.outer_ipv4
	mov h.outer_ipv4.total_len h.ipv4.total_len
	add h.outer_ipv4.total_len 0x32
	mov h.outer_ipv4.version_ihl 0x45
	mov h.outer_ipv4.dscp_ecn 0x0
	mov h.outer_ipv4.identification 0x0
	mov h.outer_ipv4.flags_frag_offset 0x4000
	mov h.outer_ipv4.ttl 0x40
	mov h.outer_ipv4.protocol 0x11
	mov h.outer_ipv4.header_checksum 0x0
	mov h.outer_ipv4.src_addr t.src_addr
	mov h.outer_ipv4.dst_addr t.dst_addr
	mov h.cksum_state.state_0 0x0
	ckadd h.cksum_state.state_0 h.outer_ipv4
	mov h.outer_ipv4.header_checksum h.cksum_state.state_0
	validate h.outer_udp
	mov h.outer_udp.hdr_length h.ipv4.total_len
	add h.outer_udp.hdr_length 0x1e
	mov h.outer_udp.src_port m.local_metadata__hash10
	mov h.outer_udp.dst_port t.dst_port
	mov h.outer_udp.checksum 0x0
	validate h.vxlan
	mov h.vxlan.flags 0x8
	mov h.vxlan.reserved 0x0
	mov h.vxlan.vni t.vni
	mov h.vxlan.reserved2 0x0
	validate h.outer_ethernet
	mov h.outer_ethernet.ether_type 0x800
	return
}

action set_src_mac args instanceof set_src_mac_arg_t {
	mov h.outer_ethernet.src_addr t.src_mac_addr
	return
}

action set_outer_mac args instanceof set_outer_mac_arg_t {
	mov h.outer_ethernet.dst_addr t.dst_mac_addr
	return
}

action decap_outer_ipv4 args instanceof decap_outer_ipv4_arg_t {
	mov m.local_metadata__tunnel_id12 t.tunnel_id
	or m.MainControlT_vendormeta_mod_action_ref 0x4
	return
}

action set_tunnel args instanceof set_tunnel_arg_t {
	mov m.MainControlT_vendormeta_mod_data_ptr t.tunnel_id
	mov m.local_metadata__ipv4_dst_match11 t.dst_addr
	mov m.local_metadata__is_tunnel9 0x1
	return
}

action l2_fwd args instanceof l2_fwd_arg_t {
	mov m.pna_main_output_metadata_output_port t.port
	return
}

action l2_fwd_1 args instanceof l2_fwd_1_arg_t {
	mov m.pna_main_output_metadata_output_port t.port
	return
}

action l2_fwd_2 args instanceof l2_fwd_2_arg_t {
	mov m.pna_main_output_metadata_output_port t.port
	return
}

action set_nexthop args instanceof set_nexthop_arg_t {
	or m.MainControlT_vendormeta_mod_action_ref 0x8
	mov m.MainControlT_vendormeta_neighbor_mod_data_ptr t.neighbor_id
	mov m.local_metadata__rif_mod_map_id7 t.router_interface_id
	mov m.pna_main_output_metadata_output_port t.egress_port
	return
}

action set_nexthop_id args instanceof set_nexthop_id_arg_t {
	mov m.local_metadata__nexthop_id6 t.nexthop_id
	return
}

action ipv4_table_set_group_id args instanceof ipv4_table_set_group_id_arg_t {
	mov m.MainControlT_as_ecmp_group_id t.group_id
	return
}

action ipv4_table_set_member_id args instanceof ipv4_table_set_member_id_arg_t {
	mov m.MainControlT_as_ecmp_member_id t.member_id
	return
}

action set_control_dest args instanceof set_control_dest_arg_t {
	mov m.pna_main_output_metadata_output_port t.port_id
	return
}

action set_control_dest_1 args instanceof set_control_dest_1_arg_t {
	mov m.pna_main_output_metadata_output_port t.port_id
	return
}

action push_vlan_fwd args instanceof push_vlan_fwd_arg_t {
	validate h.outer_vlan_0
	mov h.outer_vlan_0.ether_type h.outer_ethernet.ether_type
	mov h.outer_vlan_0.pcp_cfi_vid t.vlan_tag
	mov h.outer_ethernet.ether_type 0x8100
	mov m.pna_main_output_metadata_output_port t.port
	return
}

action pop_vlan_fwd args instanceof pop_vlan_fwd_arg_t {
	mov h.outer_ethernet.ether_type h.outer_vlan_0.ether_type
	invalidate h.outer_vlan_0
	mov m.pna_main_output_metadata_output_port t.port
	return
}

table vxlan_encap_mod_table {
	key {
		m.MainControlT_vendormeta_mod_data_ptr exact
	}
	actions {
		vxlan_encap
		NoAction
	}
	default_action NoAction args none const
	size 0x10000
}


table rif_mod_table {
	key {
		m.local_metadata__rif_mod_map_id7 exact
	}
	actions {
		set_src_mac
		NoAction
	}
	default_action NoAction args none const
	size 0x200
}


table neighbor_mod_table {
	key {
		m.MainControlT_vendormeta_neighbor_mod_data_ptr exact
	}
	actions {
		set_outer_mac
		NoAction
	}
	default_action NoAction args none const
	size 0x10000
}


table ipv4_tunnel_term_table {
	key {
		m.local_metadata__tunnel_tun_type14 exact
		m.linux_networking_control_ipv4_tunnel_term_table_outer_ipv4_0 exact
		m.linux_networking_control_ipv4_tunnel_term_table_outer_ipv4_1 exact
	}
	actions {
		decap_outer_ipv4
		drop
	}
	default_action drop args none 
	size 0x10000
}


table l2_fwd_rx_table {
	key {
		h.outer_ethernet.dst_addr exact
	}
	actions {
		l2_fwd
		drop_1
	}
	default_action drop_1 args none const
	size 0x10000
}


table l2_fwd_rx_with_tunnel_table {
	key {
		h.ethernet.dst_addr exact
	}
	actions {
		l2_fwd_1
		drop_2
	}
	default_action drop_2 args none const
	size 0x10000
}


table l2_fwd_tx_table {
	key {
		h.outer_ethernet.dst_addr exact
	}
	actions {
		l2_fwd_2
		set_tunnel
		drop_3
	}
	default_action drop_3 args none const
	size 0x10000
}


table nexthop_table {
	key {
		m.local_metadata__nexthop_id6 exact
	}
	actions {
		set_nexthop
		drop_4
	}
	default_action drop_4 args none const
	size 0x10000
}


table ipv4_table {
	key {
		m.local_metadata__ipv4_dst_match11 lpm
	}
	actions {
		ipv4_table_set_group_id
		ipv4_table_set_member_id
		NoAction
	}
	default_action NoAction args none 
	size 0x10000
}


table as_ecmp {
	key {
		m.MainControlT_as_ecmp_member_id exact
	}
	actions {
		set_nexthop_id
		drop_5
		NoAction
	}
	default_action NoAction args none 
	size 0x10000
}


table handle_rx_control_pkts_table {
	key {
		m.pna_main_input_metadata_input_port exact
	}
	actions {
		set_control_dest
		drop_6
	}
	default_action drop_6 args none const
	size 0x10000
}


table handle_tx_control_vlan_pkts_table {
	key {
		m.pna_main_input_metadata_input_port exact
		m.local_metadata__vlan_id8 exact
	}
	actions {
		pop_vlan_fwd
		drop_7
	}
	default_action drop_7 args none const
	size 0x10000
}


table handle_tx_control_pkts_table {
	key {
		m.pna_main_input_metadata_input_port exact
	}
	actions {
		push_vlan_fwd
		set_control_dest_1
		drop_8
	}
	default_action drop_8 args none const
	size 0x10000
}


selector as_ecmp_sel {
	group_id m.MainControlT_as_ecmp_group_id
	selector {
		m.linux_networking_control_as_ecmp_sel_outer_ipv4_src_addr
		m.linux_networking_control_as_ecmp_sel_outer_ipv4_dst_addr
		m.linux_networking_control_as_ecmp_sel_outer_ipv4_protocol
		m.linux_networking_control_as_ecmp_sel_outer_udp_src_port
		m.linux_networking_control_as_ecmp_sel_outer_udp_dst_port
	}
	member_id m.MainControlT_as_ecmp_member_id
	n_groups_max 128
	n_members_per_group_max 1024
}

apply {
	rx m.pna_main_input_metadata_input_port
	extract h.outer_ethernet
	mov m.local_metadata__control_packet1 0x0
	jmpeq PACKET_PARSER_PARSE_OUTER_VLAN_0 h.outer_ethernet.ether_type 0x8100
	jmpeq PACKET_PARSER_PARSE_OUTER_ARP h.outer_ethernet.ether_type 0x806
	jmpeq PACKET_PARSER_PARSE_OUTER_IPV4 h.outer_ethernet.ether_type 0x800
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_OUTER_VLAN_0 :	extract h.outer_vlan_0
	mov m.local_metadata__vlan_id8 h.outer_vlan_0.pcp_cfi_vid
	jmpeq PACKET_PARSER_PARSE_OUTER_VLAN_1 h.outer_vlan_0.ether_type 0x8100
	jmpeq PACKET_PARSER_PARSE_OUTER_ARP h.outer_vlan_0.ether_type 0x806
	jmpeq PACKET_PARSER_PARSE_OUTER_IPV4 h.outer_vlan_0.ether_type 0x800
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_OUTER_VLAN_1 :	extract h.outer_vlan_1
	mov m.local_metadata__vlan_id8 h.outer_vlan_1.pcp_cfi_vid
	jmpeq PACKET_PARSER_PARSE_OUTER_ARP h.outer_vlan_1.ether_type 0x806
	jmpeq PACKET_PARSER_PARSE_OUTER_IPV4 h.outer_vlan_1.ether_type 0x800
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_OUTER_IPV4 :	extract h.outer_ipv4
	jmpeq PACKET_PARSER_PARSE_OUTER_ICMP h.outer_ipv4.protocol 0x1
	jmpeq PACKET_PARSER_PARSE_OUTER_TCP h.outer_ipv4.protocol 0x6
	jmpeq PACKET_PARSER_PARSE_OUTER_UDP h.outer_ipv4.protocol 0x11
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_OUTER_UDP :	extract h.outer_udp
	jmpeq PACKET_PARSER_PARSE_VXLAN h.outer_udp.dst_port 0x12b5
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_VXLAN :	extract h.vxlan
	mov m.local_metadata__tunnel_tun_type14 0x2
	extract h.ethernet
	jmpeq PACKET_PARSER_PARSE_VLAN_0 h.ethernet.ether_type 0x8100
	jmpeq PACKET_PARSER_PARSE_ARP h.ethernet.ether_type 0x806
	jmpeq PACKET_PARSER_PARSE_IPV4 h.ethernet.ether_type 0x800
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_VLAN_0 :	extract h.vlan_0
	mov m.local_metadata__vlan_id8 h.vlan_0.pcp_cfi_vid
	jmpeq PACKET_PARSER_PARSE_VLAN_1 h.vlan_0.ether_type 0x8100
	jmpeq PACKET_PARSER_PARSE_ARP h.vlan_0.ether_type 0x806
	jmpeq PACKET_PARSER_PARSE_IPV4 h.vlan_0.ether_type 0x800
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_VLAN_1 :	extract h.vlan_1
	mov m.local_metadata__vlan_id8 h.vlan_1.pcp_cfi_vid
	jmpeq PACKET_PARSER_PARSE_ARP h.vlan_1.ether_type 0x806
	jmpeq PACKET_PARSER_PARSE_IPV4 h.vlan_1.ether_type 0x800
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_IPV4 :	extract h.ipv4
	jmpeq PACKET_PARSER_PARSE_ICMP h.ipv4.protocol 0x1
	jmpeq PACKET_PARSER_PARSE_TCP h.ipv4.protocol 0x6
	jmpeq PACKET_PARSER_PARSE_UDP h.ipv4.protocol 0x11
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_UDP :	extract h.udp
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_TCP :	extract h.tcp
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_ICMP :	extract h.icmp
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_ARP :	extract h.arp
	mov m.local_metadata__control_packet1 0x1
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_OUTER_TCP :	extract h.outer_tcp
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_OUTER_ICMP :	extract h.outer_icmp
	jmp PACKET_PARSER_ACCEPT
	PACKET_PARSER_PARSE_OUTER_ARP :	extract h.outer_arp
	mov m.local_metadata__control_packet1 0x1
	PACKET_PARSER_ACCEPT :	mov m.MainControlT_vendormeta_mod_action_ref 0x0
	mov m.MainControlT_vendormeta_mod_data_ptr 0xffff
	regrd m.pna_main_input_metadata_direction direction m.pna_main_input_metadata_input_port
	jmpneq LABEL_FALSE m.pna_main_input_metadata_direction 0x0
	jmpneq LABEL_FALSE_0 m.local_metadata__control_packet1 0x1
	table handle_rx_control_pkts_table
	jmp LABEL_END
	LABEL_FALSE_0 :	jmpnv LABEL_FALSE_1 h.outer_ipv4
	jmpnv LABEL_FALSE_1 h.ethernet
	mov m.linux_networking_control_ipv4_tunnel_term_table_outer_ipv4_0 h.outer_ipv4.src_addr
	mov m.linux_networking_control_ipv4_tunnel_term_table_outer_ipv4_1 h.outer_ipv4.dst_addr
	table ipv4_tunnel_term_table
	table l2_fwd_rx_with_tunnel_table
	jmp LABEL_END
	LABEL_FALSE_1 :	table l2_fwd_rx_table
	jmp LABEL_END
	LABEL_FALSE :	jmpneq LABEL_END m.pna_main_input_metadata_direction 0x1
	jmpneq LABEL_FALSE_3 m.local_metadata__control_packet1 0x1
	jmpnv LABEL_FALSE_3 h.outer_vlan_0
	table handle_tx_control_vlan_pkts_table
	jmp LABEL_END
	LABEL_FALSE_3 :	jmpneq LABEL_FALSE_4 m.local_metadata__control_packet1 0x1
	table handle_tx_control_pkts_table
	jmp LABEL_END
	LABEL_FALSE_4 :	table l2_fwd_tx_table
	jmpa LABEL_SWITCH set_tunnel
	jmp LABEL_END
	LABEL_SWITCH :	mov m.MainControlT_tmp_1 h.outer_ethernet.src_addr
	mov m.MainControlT_tmp_2 h.outer_ethernet.dst_addr
	hash jhash m.local_metadata__hash10  m.MainControlT_tmp_1 m.MainControlT_tmp_2
	mov m.MainControlT_as_ecmp_member_id 0x0
	mov m.MainControlT_as_ecmp_group_id 0x0
	table ipv4_table
	mov m.linux_networking_control_as_ecmp_sel_outer_ipv4_src_addr h.outer_ipv4.src_addr
	mov m.linux_networking_control_as_ecmp_sel_outer_ipv4_dst_addr h.outer_ipv4.dst_addr
	mov m.linux_networking_control_as_ecmp_sel_outer_ipv4_protocol h.outer_ipv4.protocol
	mov m.linux_networking_control_as_ecmp_sel_outer_udp_src_port h.outer_udp.src_port
	mov m.linux_networking_control_as_ecmp_sel_outer_udp_dst_port h.outer_udp.dst_port
	table as_ecmp_sel
	table as_ecmp
	table vxlan_encap_mod_table
	table nexthop_table
	LABEL_END :	mov m.MainControlT_tmp m.MainControlT_vendormeta_mod_action_ref
	and m.MainControlT_tmp 0x4
	jmpeq LABEL_END_5 m.MainControlT_tmp 0x0
	invalidate h.outer_ethernet
	invalidate h.outer_ipv4
	invalidate h.outer_udp
	invalidate h.vxlan
	LABEL_END_5 :	mov m.MainControlT_tmp_0 m.MainControlT_vendormeta_mod_action_ref
	and m.MainControlT_tmp_0 0x8
	jmpeq LABEL_END_6 m.MainControlT_tmp_0 0x0
	mov m.MainControlT_vendormeta_mod_data_ptr 0xffff
	table neighbor_mod_table
	jmpa LABEL_SWITCH_0 set_outer_mac
	jmp LABEL_END_6
	LABEL_SWITCH_0 :	table rif_mod_table
	LABEL_END_6 :	emit h.outer_ethernet
	emit h.outer_vlan_0
	emit h.outer_vlan_1
	emit h.outer_arp
	emit h.outer_ipv4
	emit h.outer_icmp
	emit h.outer_udp
	emit h.outer_tcp
	emit h.vxlan
	emit h.ethernet
	emit h.vlan_0
	emit h.vlan_1
	emit h.arp
	emit h.ipv4
	emit h.icmp
	emit h.tcp
	emit h.udp
	tx m.pna_main_output_metadata_output_port
}


