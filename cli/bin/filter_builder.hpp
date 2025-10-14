// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#pragma once

#include <optional>
#include <cxxopts.hpp>
#include <etfilters.hpp>

std::optional<etcpp::FilterCriteria> buildFilterCriteria(const cxxopts::ParseResult& argParseResult);