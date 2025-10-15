#
# Shared CPP library
#

find_package(fmt CONFIG REQUIRED)
find_package(OpenSSL REQUIRED)
find_package(nlohmann_json CONFIG REQUIRED)

find_package(Threads REQUIRED)

add_library(etcpp STATIC include/etsession.hpp)

target_sources(etcpp PRIVATE
    include/etsession.hpp
    include/etbackup.hpp
    include/etrestore.hpp
    include/etlog.hpp
    include/etutil.hpp
    include/et.hpp
    include/etencryption.hpp
    include/etfilters.hpp
    include/etincremental.hpp
    lib/etsession.cpp
    lib/etbackup.cpp
    lib/etrestore.cpp
    lib/etlog.cpp
    lib/etutil.cpp
    lib/et.cpp
    lib/etencryption.cpp
    lib/etfilters.cpp
    lib/etincremental.cpp
)

if (UNIX)
    target_sources(etcpp PRIVATE lib/etutil_unix.cpp)
endif()

if (WIN32)
    target_sources(etcpp PRIVATE lib/etutil_win32.cpp)
elseif(APPLE)
    target_sources(etcpp PRIVATE lib/etutil_darwin.cpp)
elseif(UNIX)
    target_sources(etcpp PRIVATE lib/etutil_linux.cpp)
else()
    message(FATAL_ERROR "Unknown platform")
endif()

set(config_file_dir "${CMAKE_CURRENT_BINARY_DIR}/config")
set(config_file "${config_file_dir}/etconfig.hpp")
set(config_file_in "${CMAKE_CURRENT_SOURCE_DIR}/cmake/etconfig.hpp.in")

configure_file("${config_file_in}" "${config_file}" @ONLY)

target_link_libraries(etcpp
        PRIVATE etcore
        PUBLIC fmt::fmt Threads::Threads OpenSSL::SSL OpenSSL::Crypto nlohmann_json::nlohmann_json
)

if (WIN32)
    target_link_libraries(etcpp PUBLIC Userenv)
endif()

target_include_directories(etcpp PUBLIC include ${config_file_dir})

apply_cpp_flags(etcpp)

add_subdirectory(tests)
