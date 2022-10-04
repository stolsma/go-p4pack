#!/bin/bash
##
## Copyright(c) 2021 Intel Corporation.
## Copyright(c) 2022 Sander Tolsma.
##
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
##
## http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.
##

apply_patch()
{
	local PATCH_FILES=(007-Fix-pipeline-structs-initialization.patch)

	CURRENT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
	readonly DEFAULT_DPDK_HOME="${CURRENT_PATH}"/dpdk
	DPDK_SRC_PATH=${DPDK_HOME:-$DEFAULT_DPDK_HOME}
	DPDK_PATCH_PATH="${CURRENT_PATH}"patches
	DPDK_SRC_GIT_FILE="${DPDK_SRC_PATH}"/.git

	echo *******************************************************************************
	echo Patch script:
	echo "DPDK Home             : $DPDK_HOME"
	echo "Patch DPDK SRC path   : $DPDK_SRC_PATH"
	echo "Patch SRC path        : $DPDK_PATCH_PATH"
	echo "Src .git file         : $DPDK_SRC_GIT_FILE"
	echo

	# Skip patching if branch is already compiled (i.e. .../build dir exists).
	if [ -d "${DPDK_SRC_PATH}/build" ]; then
		exit 1;
	fi

	# Need to check .git in dpdk_src. There are some cases where we remove .git after patch
	# apply and start the compilation.
	if [ -e "${DPDK_SRC_GIT_FILE}" ]; then
		# Let's clean all the changes
		(cd "${DPDK_SRC_PATH}" || exit; git checkout ./*)

		# Validate and apply the patch
		for i in "${PATCH_FILES[@]}"; do
			echo "Try to apply: ${DPDK_PATCH_PATH}/${i}"
			if [ -e "${DPDK_PATCH_PATH}/${i}" ]; then
				cd "${DPDK_SRC_PATH}" || exit;
				git apply "${DPDK_PATCH_PATH}/${i}";
				echo "Applied: ${DPDK_PATCH_PATH}/${i}"
				echo
			fi
		done
	else
		# This is special case where we don't have .git dir. For example downloaded and extracted source files
		# from tar. As a workaround we have to apply the patch using <patch -p1>

		# Validate and apply the patch
		for i in "${PATCH_FILES[@]}"; do
			echo "Try to apply: ${DPDK_PATCH_PATH}/${i}"
			if [ -e "${DPDK_PATCH_PATH}/${i}" ]; then
				cd "${DPDK_SRC_PATH}" || exit;
				patch -p1 < "${DPDK_PATCH_PATH}/${i}";
				echo "Applied: ${DPDK_PATCH_PATH}/${i}"
				echo
			fi
		done
	fi
	echo *******************************************************************************
}

apply_patch
