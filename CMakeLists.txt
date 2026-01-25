cmake_minimum_required(VERSION 3.30)
project(apm)

set(CMAKE_CXX_STANDARD 17)

include_directories(.)
include_directories(google.com/abseil-cpp)
include_directories(google.com/webrtc)

file(GLOB_RECURSE ABSEIL_SOURCES
        "google.com/abseil-cpp/*.cc"
        "google.com/abseil-cpp/*.c"
        "google.com/abseil-cpp/*.cpp"
)

add_library(abseil STATIC ${ABSEIL_SOURCES})
target_link_libraries(abseil PRIVATE "-framework CoreFoundation")

file(GLOB_RECURSE WEBRTC_SOURCES
        "google.com/webrtc/*.cc"
        "google.com/webrtc/*.c"
        "google.com/webrtc/*.m"
        "google.com/webrtc/*.S"
        "google.com/webrtc/*.cpp"
)
add_library(webrtc STATIC ${WEBRTC_SOURCES})

target_compile_definitions(webrtc
        PRIVATE
        WEBRTC_POSIX
        WEBRTC_MAC
        WEBRTC_APM_DEBUG_DUMP=0
)
target_link_libraries(webrtc PRIVATE abseil)

add_library(bridge bridge.h bridge.cpp)
target_link_libraries(bridge PRIVATE webrtc)

add_executable(apm main.cpp)
target_link_libraries(apm PRIVATE bridge)
