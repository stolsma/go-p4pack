# Go-P4Pack DPDK SWX Package

THis package supports the DPDK SWX P4 Spec programmable dataplane.

## Supported DPDK SWX C functions

Crossed functions are supported by this package. All functions are tagged 'experimental' in DPDK at the moment.

### added in DPDK 20.11

- [x] rte_swx_ctl_action_arg_info_get;
- [x] rte_swx_ctl_action_info_get;
- [x] rte_swx_ctl_pipeline_abort;
- [x] rte_swx_ctl_pipeline_commit;
- [x] rte_swx_ctl_pipeline_create;
- [x] rte_swx_ctl_pipeline_free;
- [x] rte_swx_ctl_pipeline_info_get;
- [ ] rte_swx_ctl_pipeline_mirroring_session_set;
- [x] rte_swx_ctl_pipeline_numa_node_get;
- [x] rte_swx_ctl_pipeline_port_in_stats_read;
- [x] rte_swx_ctl_pipeline_port_out_stats_read;
- [x] rte_swx_ctl_pipeline_table_default_entry_add;
- [x] rte_swx_ctl_pipeline_table_entry_add;
- [x] rte_swx_ctl_pipeline_table_entry_delete;
- [x] rte_swx_ctl_pipeline_table_entry_read;
- [ ] rte_swx_ctl_pipeline_table_fprintf;
- [x] rte_swx_ctl_table_action_info_get;
- [x] rte_swx_ctl_table_info_get;
- [x] rte_swx_ctl_table_match_field_info_get;
- [ ] rte_swx_ctl_table_ops_get;
- [ ] rte_swx_pipeline_action_config;
- [ ] rte_swx_pipeline_build;
- [x] rte_swx_pipeline_build_from_spec;
- [x] rte_swx_pipeline_config; (implemented in pipeline.Init())
- [ ] rte_swx_pipeline_extern_func_register;
- [ ] rte_swx_pipeline_extern_object_config;
- [ ] rte_swx_pipeline_extern_type_member_func_register;
- [ ] rte_swx_pipeline_extern_type_register;
- [ ] rte_swx_pipeline_flush;
- [x] rte_swx_pipeline_free;
- [ ] rte_swx_pipeline_instructions_config;
- [ ] rte_swx_pipeline_mirroring_config;
- [ ] rte_swx_pipeline_packet_header_register;
- [ ] rte_swx_pipeline_packet_metadata_register;
- [x] rte_swx_pipeline_port_in_config;
- [ ] rte_swx_pipeline_port_in_type_register;
- [x] rte_swx_pipeline_port_out_config;
- [ ] rte_swx_pipeline_port_out_type_register;
- [x] rte_swx_pipeline_run; (implemented in Thread C code)
- [ ] rte_swx_pipeline_struct_type_register;
- [ ] rte_swx_pipeline_table_config;
- [ ] rte_swx_pipeline_table_state_get;
- [ ] rte_swx_pipeline_table_state_set;
- [ ] rte_swx_pipeline_table_type_register;

### added in DPDK 21.05

- [ ] rte_swx_ctl_metarray_info_get;
- [ ] rte_swx_ctl_meter_profile_add;
- [ ] rte_swx_ctl_meter_profile_delete;
- [ ] rte_swx_ctl_meter_reset;
- [ ] rte_swx_ctl_meter_set;
- [ ] rte_swx_ctl_meter_stats_read;
- [ ] rte_swx_ctl_pipeline_regarray_read;
- [ ] rte_swx_ctl_pipeline_regarray_write;
- [ ] rte_swx_ctl_pipeline_table_stats_read;
- [ ] rte_swx_ctl_regarray_info_get;
- [ ] rte_swx_pipeline_metarray_config;
- [ ] rte_swx_pipeline_regarray_config;

### added in DPDK 21.08

- [ ] rte_swx_pipeline_selector_config;
- [ ] rte_swx_ctl_pipeline_selector_fprintf;
- [ ] rte_swx_ctl_pipeline_selector_group_add;
- [ ] rte_swx_ctl_pipeline_selector_group_delete;
- [ ] rte_swx_ctl_pipeline_selector_group_member_add;
- [ ] rte_swx_ctl_pipeline_selector_group_member_delete;
- [ ] rte_swx_ctl_pipeline_selector_stats_read;
- [ ] rte_swx_ctl_selector_info_get;
- [ ] rte_swx_ctl_selector_field_info_get;
- [ ] rte_swx_ctl_selector_group_id_field_info_get;
- [ ] rte_swx_ctl_selector_member_id_field_info_get;

### added in DPDK 21.11

- [*] rte_swx_ctl_pipeline_learner_default_entry_add;
- [*] rte_swx_ctl_pipeline_learner_default_entry_read;
- [ ] rte_swx_ctl_pipeline_learner_stats_read;
- [ ] rte_swx_ctl_learner_action_info_get;
- [ ] rte_swx_ctl_learner_info_get;
- [ ] rte_swx_ctl_learner_match_field_info_get;
- [ ] rte_swx_pipeline_learner_config;

### added in DPDK 22.07

- [ ] rte_swx_ctl_pipeline_learner_timeout_get;
- [ ] rte_swx_ctl_pipeline_learner_timeout_set;
- [ ] rte_swx_pipeline_hash_func_register;
