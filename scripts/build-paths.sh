#!/usr/bin/env bash
# Canonical local build output layout for QA, deployment and release candidates:
#   build/android/  staged Android APK plus its bundled seeded demo data
#   build/ios/      staged iOS demo assets and, with Xcode, the simulator .app
#   build/web/      static web release bundle
#   build/server/   local server executables
# Source this from build helpers; run directly to create all four directories.

J3_ROOT_DIR="${J3_ROOT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
J3_BUILD_BASE="${J3_BUILD_BASE:-$J3_ROOT_DIR/build}"
J3_BUILD_ANDROID="$J3_BUILD_BASE/android"
J3_BUILD_IOS="$J3_BUILD_BASE/ios"
J3_BUILD_WEB="$J3_BUILD_BASE/web"
J3_BUILD_SERVER="$J3_BUILD_BASE/server"

j3_prepare_build_dir() {
  mkdir -p "$1"
}

j3_prepare_build_dirs() {
  j3_prepare_build_dir "$J3_BUILD_ANDROID"
  j3_prepare_build_dir "$J3_BUILD_IOS"
  j3_prepare_build_dir "$J3_BUILD_WEB"
  j3_prepare_build_dir "$J3_BUILD_SERVER"
}

j3_stage_dir() {
  local src="$1" dest="$2"
  # Replace rather than merge so successive flavour builds cannot leave stale
  # hashed chunks from a previous bundle in the staged output.
  rm -rf "$dest"
  j3_prepare_build_dir "$dest"
  cp -R "$src"/. "$dest"/
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  j3_prepare_build_dirs
  echo "✅ Local build directories ready under $J3_BUILD_BASE"
  printf '   %s\n' "$J3_BUILD_ANDROID" "$J3_BUILD_IOS" "$J3_BUILD_WEB" "$J3_BUILD_SERVER"
fi
