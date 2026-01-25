src=/Users/liuhongliang/GolandProjects/webrtc-sdk/webrtc/$1
dst=./google.com/webrtc/$1
echo $src
echo $dst
src_dir=`dirname $src`
dir=`dirname $dst`
mkdir -p $dir
cp $src $dst
src_cpp=$src_dir/`basename $src .h`.cc
dst_cpp=$dir/`basename $dst .h`.cc
echo $src_cpp
echo $dst_cpp
cp $src_cpp $dst_cpp
src_c=$src_dir/`basename $src .h`.c
dst_c=$dir/`basename $dst .h`.c
echo $src_c
echo $dst_c
cp $src_c $dst_c
