#!/usr/bin/env bash

# TODO
#
# √ inside core [command]
# √ Stop Hermes
# √ h start|stop
# √ Restart Hermes
# √ h status
# √ aliases: st -> status, hermes -> h, etc.
# √ Error when tmp/pids/server.pid not found.
# Onboarding support
# Self install via wget | bash
# h morning (sync all 4 repos)
# √ Support --verbose
# √ run core not in daemon mode
# Help doc
# Colorize outputs
# √ Reorganize directory hierarchy
# Auto-completion
# Support --dry-run
# h start default to run current service first

HERMES_HOME=${HERMES_HOME:-$HOME/code/wrapports/hermes}

IMAGE_SERVICE_HOME=$HERMES_HOME/services/image_service/
CORE_HOME=$HERMES_HOME/core/
FRONTEND_HOME=$HERMES_HOME/frontend/

COMMON_HOME=$HERMES_HOME/hermes_common/

SUBURBAN_HOME=$HERMES_HOME/../hermes_suburban/
COMMUNITY_HOME=$HERMES_HOME/../hermes_community/
WEDDINGS_HOME=$HERMES_HOME/../hermes_weddings/

COMPONENTS="mongodb image_service core frontend"

debug() {
  echo "# DEBUG ############" "$*"
}

error() {
  echo "#[Hermes][E]: $*" 1>&2
  return 1
}

warn() {
  echo "#[Hermes][W]: $*" 1>&2
}

info() {
  echo -e "#[Hermes][I]: $*"
}

silent() {
  if [[ -n "$HERMES_VERBOSE" && "$HERMES_VERBOSE" != "0" && "$HERMES_VERBOSE" != "false" ]]; then
    ( $* )
  else
    ( $* ) &>/dev/null
  fi
}

_hermes_cd() {
  cd $* &>/dev/null
}

hermes_home() {
  _hermes_cd $HERMES_HOME
}

hermes_image_service() {
  _hermes_cd $IMAGE_SERVICE_HOME
}

hermes_core() {
  _hermes_cd $CORE_HOME
}

hermes_frontend() {
  _hermes_cd $FRONTEND_HOME
}

hermes_common() {
  _hermes_cd $COMMON_HOME
}

hermes_suburban() {
  _hermes_cd $SUBURBAN_HOME
}

hermes_community() {
  _hermes_cd $COMMUNITY_HOME
}

hermes_weddings() {
  _hermes_cd $WEDDINGS_HOME
}

inside() {
  local component=${1}; shift
  local run_command=${*}
  local old_pwd=$(pwd)

  h $component
  $run_command
  local result=$?
  _hermes_cd $old_pwd

  return $result
}

pid_of() {
  local name=${1:?Component needs to be specified.}
  local pid=$(inside "${name}" 'cat tmp/pids/server.pid')
  echo ${pid:-Not Found}
}

is_mongodb_running() {
  ps aux | grep --color '[m]ongod\b'
}

is_rack_based_app_running() {
  local pid=$(cat tmp/pids/server.pid 2>/dev/null)
  ps aux | grep --color "${pid:=Not Found}.*[r]uby"
}

is_image_service_running() {
  inside 'image_service' 'is_rack_based_app_running'
}

is_core_running() {
  inside 'core' 'is_rack_based_app_running'
}

is_frontend_running() {
  inside 'frontend' 'is_rack_based_app_running'
}

start_sinatra_app() {
  local name=${1:-unnamed}
  local daemon_flag=''
  if [[ "$DAEMONIZE" == "true" ]]; then
    daemon_flag='-D'
    info "Starting ${component} as a Daemon..."
  else
    info "Starting ${component} in the foreground..."
  fi

  mkdir -p tmp/pids
  silent 'bundle' && silent "rackup ${daemon_flag} -P tmp/pids/server.pid config.ru" || warn "Start Sinatra app ${name} FAILED!"
}

start_rails_app() {
  local name=${1:-unnamed}
  local daemon_flag=''
  if [[ "$DAEMONIZE" == "true" ]]; then
    daemon_flag='-d'
    info "Starting ${component} as a Daemon..."
  else
    info "Starting ${component} in the foreground..."
  fi

  mkdir -p tmp/pids
  silent 'bundle' && silent "rails server ${daemon_flag}" || warn "Start Rails app ${name} FAILED!"
}

stop_sinatra_app() {
  local name=${1:?Component needs to be specified.}
  silent "kill -9 $(pid_of ${name})" || warn 'Stop Sinatra app ${name} FAILED.'
}

stop_rails_app() {
  local name=${1:?Component needs to be specified.}
  silent "kill -9 $(pid_of ${name})" || warn 'Stop Rails app ${name} FAILED.'
}

hermes_start_mongodb() {
  info 'Starting Database...'
  silent 'brew services start mongodb' || warn 'MongoDB is already running!'
}

hermes_start_image_service() {
  silent 'is_image_service_running' && warn 'Service "image_service" is already running!' && return 0
  inside 'image_service' 'start_sinatra_app image_service'
}

hermes_start_core() {
  silent 'is_core_running' && warn 'Application "core" is already running!' && return 0
  inside 'core' 'start_rails_app core'
}

hermes_start_frontend() {
  silent 'is_frontend_running' && warn 'Application "frontend" is already running!' && return 0
  inside 'frontend' 'start_rails_app frontend'
}

hermes_start() {
  DAEMONIZE="true"
  for component in ${*:-$COMPONENTS}; do
    hermes_start_${component}
  done
}

hermes_run() {
  DAEMONIZE="false"
  local component=$1
  hermes_start_${component}
}

hermes_stop_mongodb() {
  silent 'brew services stop mongodb' || warn 'MongoDB is not running.'
}

hermes_stop_image_service() {
  ( silent 'is_image_service_running' && inside 'image_service' 'stop_sinatra_app image_service' ) || \
    warn 'Service "image_service" is not running.'
}

hermes_stop_core() {
  ( silent 'is_core_running' && inside 'core' 'stop_rails_app core' ) || \
    warn 'Application "core" is not running.'
}

hermes_stop_frontend() {
  ( silent 'is_frontend_running' && inside 'frontend' 'stop_rails_app frontend' ) || \
    warn 'Application "frontend" is not running.'
}

hermes_stop() {
  for component in ${*:-$COMPONENTS}; do
    info "Stopping ${component}..."
    hermes_stop_${component};
  done
}

hermes_restart() {
  DAEMONIZE="true"
  for component in ${*:-$COMPONENTS}; do
    hermes_stop_${component} && hermes_start_${component}
  done
}

hermes_rerun() {
  DAEMONIZE="false"
  local component=$1
  hermes_stop_${component} && hermes_start_${component}
}

is_up_or_down() {
  local name=${1:?Component needs to be specified.}
  if [[ $(is_${name}_running) ]]; then echo '√'; else echo 'X'; fi
}

hermes_status() {
  for component in ${*:-$COMPONENTS}; do
    info "$(is_up_or_down ${component}) ${component}"
  done
}

# Aliases Starts
hermes_st() {
  hermes_status "$@"
}

hermes_s() {
  hermes_suburban
}

hermes_c() {
  hermes_community
}

hermes_w() {
  hermes_weddings
}

hermes_f() {
  hermes_frontend
}

hermes_start_db() {
  hermes_start_mongodb
}

hermes_stop_db() {
  hermes_stop_mongodb
}
# End of Aliases

h() {
  local sub_command=${1}; shift
  hermes_${sub_command:-home} $@ || error "Unrecognized sub-command: '$sub_command'."
}

