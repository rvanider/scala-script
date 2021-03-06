#!/bin/bash
# provides a simple include model for scala scripting

top=$1

if [ -z "$(which scala)" ]; then
  echo "scala-script requires scala to be installed"
  exit 1
fi

if [ -z "$top" ]; then
  echo "usage: scala-script script.scala [script-args]"
  echo "usage: scala-script --repl [scala-args]"
  exit 1
fi

if [ "$top" == "--repl" ]; then
  open_repl=1
elif [ ! -f "$top" ]; then
  echo "error: unable to locate $top"
  exit 1
else
  open_repl=0
fi

function debug_line
{
  if [ -n "$SCALA_SCRIPT_DEBUG" ]; then
    (>&2 echo "$*")
  fi
  return 0
}

function get_gen_file
{
  local source=$1
  # local dest="${source%.*}.g.scala"
  local dest="$(dirname $source)/.g.$(basename $source)"

  echo $dest
}

function process_file
{
  ## determine the files we will process
  #
  local file="$1"
  local padding=$2
  local parent_g_file="$3"
  local folder=$(dirname $file)
  local out_file=$(get_gen_file $file)
  debug_line "${padding}processing $(basename $out_file)"

  ## recursively process any included files
  ## in order to have them fully generated
  #
  local INCLUDES=$(cat $file | grep '//#include')
  local include
  local dirty=0
  debug_line "${padding}scanning dependencies on ${file}"
  for include in $INCLUDES; do
    if [ "//#include" != "$include" ]; then
      debug_line "${padding}found include for $include in $(basename $file)"
      local include="$folder/$include"
      local include_target=$(get_gen_file $include)
      if [ ! -f $include ]; then
        echo "error: unable to locate $include"
        exit 1
      fi
      if [ ! -f "$include_target" ]; then
        local dirty=1
        debug_line "${padding}target missing, removing parent: $(basename $parent_g_file)"
        [ -n "$parent_g_file" ] && rm -f "$parent_g_file"
      elif [ "$include" -nt "$include_target" ]; then
        local dirty=1
        debug_line "${padding}source newer, removing parent: $(basename $parent_g_file)"
        [ -n "$parent_g_file" ] && rm -f "$parent_g_file"
      fi
      process_file $include "${padding}  " "$out_file"
      local dest=$(get_gen_file $include)
      debug_line "${padding}results in $(basename $dest)"
    fi
  done

  if [ "$dirty" != "0" ]; then
    debug_line "${padding}removing generated $(basename $out_file)"
    rm -f "$out_file"
    debug_line "${padding}removing parent: $(basename $parent_g_file)"
    [ -n "$parent_g_file" ] && rm -f "$parent_g_file"
  fi

  ## final loop over the includes at our level to
  ## perform the replacement/include on the items
  ## listed - resulting in a single file with all
  ## included materials
  ##
  ## each included file is processed into a .g. file
  ## to use as a cache and for the incremental processing
  #
  local tmp_out_file="$out_file.tmp"
  local forced=0
  if [ ! -f "$out_file" ]; then
    debug_line "${padding}-- destination does not exist: $(basename $out_file)"
    local forced=1
    cp -f "$file" "$out_file"
  elif [ "$file" -nt "$out_file" ]; then
    debug_line "${padding}-- destination is older: $(basename $out_file)"
    local forced=1
    cp -f "$file" "$out_file"
  fi
  cp -fp "$out_file" "$tmp_out_file"

  if [ "$forced" == "1" ]; then
    debug_line "${padding}removing parent: $(basename $parent_g_file)"
    [ -n "$parent_g_file" ] && rm -f "$parent_g_file"
  fi
  local include
  for include in $INCLUDES; do
    if [ "//#include" != "$include" ]; then
      local short_include=$include
      local include="$folder/$include"
      debug_line "${padding}>>>[$include]"
      local g_file=$(get_gen_file $include)
      if [ "$g_file" -nt "$out_file" ] || [ $forced -ne 0 ]; then
        debug_line "${padding}>> generating $(basename $g_file) [newer than $(basename $out_file)]"
        local encoded_include=$(echo $short_include | sed -e 's|\/|\\/|g')
        debug_line "${padding}>>> [$encoded_include]"
        # {s/.*//;G;G;G;G;}
        local ws_g_file="$g_file.ws"
        # echo '' > $ws_g_file
        cat $g_file > $ws_g_file
        echo '' >> $ws_g_file
        sed -i "" \
            -e "/\/\/#include[[:space:]]*$encoded_include/r $ws_g_file" \
            -e "/\/\/#include[[:space:]]*$encoded_include/d" \
            "$tmp_out_file"
        rm -f $ws_g_file
      fi
    fi
  done
  mv "$tmp_out_file" "$out_file"
}

here=$(pwd)

if [ $open_repl -eq 1 ]; then
  folder=$here
  out_file=
else
  folder=$(dirname $top)
  file=$(basename $top)
  if [ "$folder" == "." ] || [ "$folder" == "" ]; then
    folder=$here
  fi
  out_file=$(get_gen_file "$folder/$file")
  # echo "[$here][$folder][$file][$out_file]"
  process_file "$folder/$file" "" "$out_file"
fi

## build the class path
##
if [ -d "$folder/lib" ]; then
  CP=""
  for file in $(/bin/ls -1 $folder/lib/*.jar); do
    CP="$CP:$file"
  done
else
  CP=
fi

shift
exec scala -deprecation -feature -savecompiled -classpath "$CP" "$out_file" $*
