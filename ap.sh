src=/Users/liuhongliang/IdeaProjects/abseil-cpp/$1
dst=./google.com/abseil-cpp/$1
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
